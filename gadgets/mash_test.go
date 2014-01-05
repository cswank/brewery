package brewgadgets

import (
	"fmt"
	"testing"
	"bitbucket.com/cswank/gogadgets"
)

var (
	config = &MashConfig{
		TankRadius: 7.5 * 2.54,
		ValveRadius: 0.1875 * 2.54,
		Coefficient: 0.43244,
	}
)

func TestGetVolume(t *testing.T) {
	dev, _ := NewMash(config)
	mash := dev.(*Mash)
	volume := mash.getVolume(5.0, 13.0)
	if volume != 0.36460455256750146 {
		t.Error("incorrect volume", volume)
	}
}

func TestGetDrainTime(t *testing.T) {
	dev, _ := NewMash(config)
	mash := dev.(*Mash)
	drainTime := mash.getDrainTime(5.0, 4.0)
	if drainTime != 193.4352488013887 {
		t.Error("incorrect volume", drainTime)
	}
}


func TestStart(t *testing.T) {
	dev, _ := NewMash(config)
	mash := dev.(*Mash)
	mash.targetVolume = 4.8
	out := make(chan gogadgets.Message)
	in := make(chan gogadgets.Value)
	go mash.Start(out, in)
	msg := gogadgets.Message{
		Type: "update",
		Sender: "hlt volume",
		Value: gogadgets.Value{
			Value: 5.0,
			Units: "L",
		},
	}
	out<- msg
	msg = gogadgets.Message{
		Type: "command",
		Body: "fill mash tun",
	}
	out<- msg
	msg = gogadgets.Message{
		Type: "update",
		Sender: "hlt valve",
		Value: gogadgets.Value{
		Value: true,
		},
	}
	out<- msg
	
	for {
		val := <- in
		fmt.Println(val)
		if val.Value.(float64) <= 4.8 {
			msg = gogadgets.Message{
				Type: "update",
				Sender: "hlt valve",
				Value: gogadgets.Value{
					Value: false,
				},
			}
			out<- msg
			break
		}
	}
}
