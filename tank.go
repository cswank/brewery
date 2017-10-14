package brewery

import (
	"fmt"
	"time"

	"github.com/cswank/gogadgets"
)

type Tank struct {
	master bool
	name   string
	uid    string
	out    chan<- gogadgets.Message
}

func masterTank(t *Tank) {
	t.master = true
}

func NewTank(name string, opts ...func(*Tank)) *Tank {
	t := &Tank{name: name, uid: fmt.Sprintf("%s volume", name)}
	for _, f := range opts {
		f(t)
	}
	vol.register(name, t.sendUpdate)
	return t
}

func (t *Tank) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	t.out = out
	t.sendUpdate(vol.get(t.name))
	for {
		msg := <-input
		t.readMessage(msg)
	}
}

func (t *Tank) GetDirection() string {
	return "input"
}

func (t *Tank) GetUID() string {
	return t.uid
}

func (t *Tank) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		t.sendUpdate(vol.get(t.name))
	} else if t.master {
		vol.readMessage(msg)
	}
}

func (t *Tank) sendUpdate(val float64) {
	t.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    t.uid,
		Location:  t.name,
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: val,
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}
