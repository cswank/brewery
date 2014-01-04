package brewgadgets

import (
	"testing"
	"time"
	"bitbucket.com/cswank/gogadgets"
)

type FakePoller struct {
	gogadgets.Poller
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

func TestCreate(t *testing.T) {
	poller := &FakePoller{}
	_ = &HLT{
		GPIO: poller,
		Value: 5.0,
		Units: "liters",
	}
}

func TestHLT(t *testing.T) {
	poller := &FakePoller{}
	h := &HLT{
		GPIO: poller,
		Value: 5.0,
		Units: "liters",
	}
	out := make(chan gogadgets.Message)
	in := make(chan gogadgets.Value)
	go h.Start(out, in)
	val := <-in
	if val.Value.(float64) != 5.0 {
		t.Error("should have been 5.0", val)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		out<- gogadgets.Message{
			Type: "command",
			Body: "hlt volume change",
			Value: gogadgets.Value{
				Value: -0.5,
			},
		}
	}()
	val = <-in
	if val.Value.(float64) != 4.5 {
		t.Error("should have been 4.5", val)
	}
}
