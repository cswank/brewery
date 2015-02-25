package brewgadgets

import (
	"testing"
	"time"

	"bitbucket.org/cswank/gogadgets"
)

type FakePoller struct {
	gogadgets.Poller
	trigger <-chan bool
}

func (f *FakePoller) Wait() (bool, error) {
	v := <-f.trigger
	return v, nil
}

func TestFillHLT(t *testing.T) {
	trigger := make(chan bool)
	poller := &FakePoller{
		trigger: trigger,
	}
	cfg := &BrewConfig{
		MashRadius:      20.0,
		MashValveRadius: 10.0,
		Coefficient:     0.5,
		HLTCapacity:     26.5,
	}
	k, area := getK(cfg)
	bv := BrewVolume{
		k:           k,
		mashArea:    area,
		hltFull:     make(chan bool),
		HLTCapacity: 26.5,
		poller:      poller,
		mashStop:    make(chan bool),
	}

	out := make(chan gogadgets.Message)
	in := make(chan gogadgets.Message)

	go bv.Start(out, in)

	out <- gogadgets.Message{
		Type:   "update",
		Sender: "hlt valve",
		Value: gogadgets.Value{
			Value: true,
		},
	}

	trigger <- true
	msg := <-in
	if !closeTo(msg.Value.Value.(float64), 7.0) {
		t.Error(msg.Value)
	}

	out <- gogadgets.Message{
		Type:   "update",
		Sender: "mash tun valve",
		Value: gogadgets.Value{
			Value: true,
		},
	}

	tunVolume := 0.0
	hltVolume := 0.0
	for tunVolume < 4.0 {
		msg = <-in
		if msg.Sender == "mash tun volume" {
			tunVolume = msg.Value.Value.(float64)
		} else if msg.Sender == "hlt volume" {
			hltVolume = msg.Value.Value.(float64)
		}
	}

	if !closeTo(hltVolume+tunVolume, 7.0) {
		t.Error(msg.Value)
	}

	out <- gogadgets.Message{
		Type:   "update",
		Sender: "mash tun valve",
		Value: gogadgets.Value{
			Value: false,
		},
	}

	//clear out all messages
	var stop bool
	for !stop {
		select {
		case msg = <-in:
		case <-time.After(200 * time.Millisecond):
			stop = true
		}
	}

	out <- gogadgets.Message{
		Type: "command",
		Body: "update",
	}

	v := map[string]float64{}
	for len(v) < 3 {
		msg = <-in
		v[msg.Sender] = msg.Value.Value.(float64)
	}
	if v["boiler volume"] != 0.0 {
		t.Error(v)
	}

	if !closeTo(v["hlt volume"]+v["mash tun volume"], 7.0) {
		t.Error(v)
	}
}

func closeTo(val, expected float64) bool {
	return expected-0.05 <= val && val <= expected+0.05
}
