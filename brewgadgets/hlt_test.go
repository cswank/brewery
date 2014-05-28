package brewgadgets

import (
	"bitbucket.org/cswank/gogadgets"
	"bitbucket.org/cswank/gogadgets/models"
	"bitbucket.org/cswank/gogadgets/input"
	"testing"
	"time"
)

type FakePoller struct {
	input.Poller
	val bool
}

func (f *FakePoller) Wait() (bool, error) {
	if f.val {
		time.Sleep(10 * time.Second)
	} else {
		time.Sleep(100 * time.Millisecond)
		f.val = !f.val
	}
	return f.val, nil
}

func _TestCreate(t *testing.T) {
	poller := &FakePoller{}
	_ = &HLT{
		GPIO:  poller,
		Value: 5.0,
		Units: "liters",
	}
}

func TestHLT(t *testing.T) {
	poller := &FakePoller{}
	h := &HLT{
		GPIO:  poller,
		Value: 5.0,
		Units: "liters",
	}
	out := make(chan models.Message)
	in := make(chan models.Value)
	go h.Start(out, in)
	val := <-in
	if val.Value.(float64) != 5.0 {
		t.Error("should have been 5.0", val)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		out <- models.Message{
			Type:   "update",
			Sender: "mash tun volume",
			Value: models.Value{
				Value: 0.5,
			},
		}
	}()
	val = <-in
	if val.Value.(float64) != 4.5 {
		t.Error("should have been 4.5", val)
	}
}
