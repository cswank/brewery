package brewery

import (
	"time"

	"github.com/cswank/gogadgets"
)

type Carboy struct {
	out     chan<- gogadgets.Message
	volumes map[string]float64
}

func NewCarboy() (*Carboy, error) {
	return &Carboy{
		volumes: map[string]float64{
			"boiler": 0.0,
			"carboy": 0.0,
		},
	}, nil
}

func (c *Carboy) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	c.out = out
	c.sendUpdate()
	for {
		msg := <-input
		c.readMessage(msg)
	}
}

func (c *Carboy) GetDirection() string {
	return "input"
}

func (c *Carboy) GetUID() string {
	return "carboy volume"
}

func (c *Carboy) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		c.sendUpdate()
	} else if msg.Type == "update" && msg.Sender == "boiler volume" {
		c.volumes["boiler"] = msg.Value.Value.(float64)
	} else if msg.Type == "update" && msg.Sender == "carboy pump" && msg.Value.Value == false {
		c.volumes["carboy"] = c.volumes["boiler"]
		c.sendUpdate()
	}
}

func (c *Carboy) sendUpdate() {
	c.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "carboy volume",
		Location:  "carboy",
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: c.volumes["carboy"],
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}
