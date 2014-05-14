package brewgadgets

import (
	"fmt"
	"time"
	"math/rand"
	"testing"
	"encoding/json"
	"bitbucket.org/cswank/gogadgets"
	"github.com/vaughan0/go-zmq"
)

var (
	bigConfig = &MashConfig{
		TankRadius: 7.5 * 2.54,
		ValveRadius: 1,
		Coefficient: 0.43244,
	}
)

func TestBoiler(t *testing.T) {
	b := Boiler{
		Units: "liters",
		waitTime: time.Duration(10 * time.Millisecond),
	}

	out := make(chan gogadgets.Message)
	in := make(chan gogadgets.Value)
	go b.Start(out, in)
	
	time.Sleep(10 * time.Millisecond)
	out<- gogadgets.Message{
			Type: "update",
			Sender: "mash tun volume",
			Value: gogadgets.Value{
				Value: 25.0,
			},
	}
	
	go func() {
		time.Sleep(10 * time.Millisecond)
		out<- gogadgets.Message{
			Type: "update",
			Sender: "boiler valve",
			Value: gogadgets.Value{
				Value: true,
			},
		}
	}()
	val := <-in
	
	if val.Value.(float64) != 25.0 {
		t.Error(val)
	}
}

func TestBoilerAndMash(t *testing.T) {
	mashVolume, _ := NewMash(bigConfig)
	mash := &gogadgets.Gadget{
		Location: "mash tun",
		Name: "volume",
		Input: mashVolume,
		Direction: "input",
		OnCommand: "n/a",
		OffCommand: "n/a",
		UID: "mash tun volume",
	}

	boilerVolume := &Boiler{
		Units: "liters",
		waitTime: time.Duration(1000 * time.Millisecond),
	}
	boiler := &gogadgets.Gadget{
		Location: "boiler",
		Name: "volume",
		Input: boilerVolume,
		Direction: "input",
		OnCommand: "n/a",
		OffCommand: "n/a",
		UID: "boiler volume",
	}
	pubPort := 1024 + rand.Intn(65535 - 1024)
	subPort := pubPort + 1
	
	app := &gogadgets.App{
		MasterHost: "localhost",
		PubPort: pubPort,
		SubPort: subPort,
		Gadgets: []gogadgets.GoGadget{mash, boiler},
	}
	input := make(chan gogadgets.Message)
	go app.Start(input)

	ctx, err := zmq.NewContext()
	defer ctx.Close()
	if err != nil {
		t.Fatal(err)
	}
	
	pub, err := ctx.Socket(zmq.Pub)
	defer pub.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = pub.Connect(fmt.Sprintf("tcp://localhost:%d", subPort))
	if err != nil {
		t.Fatal(err)
	}
	chans := pub.Channels()
	defer chans.Close()

	sub, err := ctx.Socket(zmq.Sub)
	defer sub.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err = sub.Connect(fmt.Sprintf("tcp://localhost:%d", pubPort)); err != nil {
		t.Fatal(err)
	}
	sub.Subscribe([]byte(""))
	
	input<- gogadgets.Message{
		Type: "update",
		Sender: "hlt volume",
		Value: gogadgets.Value{
			Value: 25.0,
			Units: "liters",
		},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		input<- gogadgets.Message{
			Type: "update",
			Sender: "mash tun valve",
			Value: gogadgets.Value{
				Value: true,
			},
		}
		time.Sleep(100 * time.Millisecond)
		input<- gogadgets.Message{
			Type: "update",
			Sender: "mash tun valve",
			Value: gogadgets.Value{
				Value: false,
			},
		}
	}()
	stepTwo := false
	for {
		parts, err := sub.Recv()
		if err != nil {
			t.Error(err)
		}
		msg := &gogadgets.Message{}
		json.Unmarshal(parts[1], msg)
		if msg.Sender == "mash tun valve" && msg.Value.Value == false {
			time.Sleep(100 * time.Millisecond)
			input<- gogadgets.Message{
				Type: "update",
				Sender: "boiler valve",
				Value: gogadgets.Value{
					Value: true,
				},
			}
			stepTwo = true
		}
		if msg.Sender == "mash tun volume" && msg.Value.Value.(float64) == 0.0 && stepTwo {
			break
		}
	}
	if !(boilerVolume.Volume > 0.0) {
		t.Error("there should be some water in the boiler", boilerVolume.Volume)
	}
	mv := mashVolume.(*Mash)
	if mv.Volume > 0.0 {
		t.Error("there should be no water in the mash tun", mv.Volume)
	}
}
