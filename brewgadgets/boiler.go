package brewgadgets

import (
	"bitbucket.org/cswank/gogadgets"
	"log"
	"time"
)

/*
Measures the volume in the boiler.

Unlike the flow of water from the HLT to the tun,
the source of the boilers water is the tun and it
is filled with grains.  The grains will screw with how
fast water flows out of the mash, so I don't try to
calculate the volume based on time (like I do for the mash
tun volume).  For this vessel, I'll simply wait a conservative
amount of time and declare that all the water from the mash
is now in the boiler.
*/

type Boiler struct {
	gogadgets.InputDevice
	Volume     float64
	Units      string
	mashVolume float64
	waitTime   time.Duration
	value      chan float64
	out        chan<- gogadgets.Value
}

func NewBoiler() (gogadgets.InputDevice, error) {
	return &Boiler{
		Units:    "L",
		waitTime: time.Duration(60 * 5 * time.Second),
	}, nil
}

func (b *Boiler) Start(in <-chan gogadgets.Message, out chan<- gogadgets.Value) {
	b.out = out
	b.value = make(chan float64)
	err := make(chan error)
	for {
		select {
		case msg := <-in:
			b.readMessage(msg)
		case val := <-b.value:
			b.Volume = val
			b.sendValue()
		case e := <-err:
			log.Println(e)
		}
	}
}

func (b *Boiler) GetValue() *gogadgets.Value {
	return &gogadgets.Value{
		Value: b.Volume,
		Units: b.Units,
	}
}

func (b *Boiler) wait(out chan<- float64) {
	time.Sleep(b.waitTime)
	totalVolume := b.mashVolume + b.Volume
	out <- totalVolume
}

func (b *Boiler) readMessage(msg gogadgets.Message) {
	if msg.Sender == "tun volume" {
		b.mashVolume = msg.Value.Value.(float64)
	} else if msg.Sender == "boiler valve" && msg.Value.Value == true {
		go b.wait(b.value)
	}
}

func (b *Boiler) sendValue() {
	b.out <- gogadgets.Value{
		Value: b.Volume,
		Units: b.Units,
	}
}
