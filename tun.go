package brewery

import (
	"time"

	"github.com/cswank/gogadgets"
)

var (
	all = []string{"hlt", "tun", "boiler", "carboy"}
)

type Tun struct {
	out chan<- gogadgets.Message
}

func NewTun() (*Tun, error) {
	return &Tun{}, nil
}

func (t *Tun) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	t.out = out
	t.sendUpdate()

	for {
		msg := <-input
		t.readMessage(msg)
	}
}

func (t *Tun) GetDirection() string {
	return "input"
}

func (t *Tun) GetUID() string {
	return "tun volume"
}

func (t *Tun) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		t.sendUpdate()
	}
}

func (t *Tun) sendUpdate() {
	t.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "tun volume",
		Location:  "tun",
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: vol.get("tun"),
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}
