package brewery

import (
	"time"

	"github.com/cswank/gogadgets"
)

type Boiler struct {
	out       chan<- gogadgets.Message
	volumes   map[string]float64
	fillTime  time.Duration
	listening bool
	stop      chan bool
}

func NewBoiler(fillTime time.Duration) (*Boiler, error) {
	return &Boiler{
		volumes: map[string]float64{
			"tun":    0.0,
			"boiler": 0.0,
		},
		fillTime: fillTime,
		stop:     make(chan bool),
	}, nil
}

func (b *Boiler) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	b.out = out
	b.sendUpdate()

	for {
		msg := <-input
		b.readMessage(msg)
	}
}

func (b *Boiler) GetDirection() string {
	return "input"
}

func (b *Boiler) GetUID() string {
	return "boiler volume"
}

func (b *Boiler) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		b.sendUpdate()
	} else if msg.Type == "update" && msg.Sender == "tun volume" {
		b.volumes["tun"] = msg.Value.Value.(float64)
	} else if msg.Type == "update" && msg.Sender == "carboy volume" {
		//if carboy declares itself full then boiler must be empty
		b.volumes["boiler"] = 0.0
		b.sendUpdate()
	} else if msg.Type == "update" && msg.Sender == "boiler valve" && msg.Value.Value == true {
		go b.updateVolume()
	} else if msg.Type == "update" && msg.Sender == "boiler valve" && msg.Value.Value == false && b.listening {
		b.stop <- true
	}
}

//The rate that water flows from the mash into the boiler isn't
//known, so wait for a safe amount of time and assume all water
//in the mash is now in the boiler.
func (b *Boiler) updateVolume() {
	b.listening = true
	for b.listening {
		select {
		case <-b.stop:
			b.listening = false
		case <-time.After(b.fillTime):
			b.volumes["boiler"] = b.volumes["tun"]
			b.sendUpdate()
			b.listening = false
		}
	}
}

func (b *Boiler) sendUpdate() {
	b.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "boiler volume",
		Location:  "boiler",
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: b.volumes["boiler"],
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}
