package brewgadgets

import (
	"time"
	"testing"
	"bitbucket.com/cswank/gogadgets"
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
