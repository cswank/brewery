package brewgadgets

import (
	"fmt"
	"time"
	"math"
	"bitbucket.com/cswank/gogadgets"
)

type Mash struct {
	gogadgets.InputDevice
	Volume float64
	previousVolume float64 
	Units string
	hltVolume float64
	out chan<- gogadgets.Value
	k float64
	tankArea float64
	valveArea float64
	valveStatus bool
	endTime time.Time
	stop chan bool
}

type MashConfig struct {
	TankRadius float64
	ValveRadius float64
	Coefficient float64
}

func NewMash(config *MashConfig) (gogadgets.InputDevice, error) {
	tankArea := math.Pi * math.Pow(config.TankRadius, 2)
	valveArea := math.Pi * math.Pow(config.ValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	k := (tankArea * x) / (valveArea * config.Coefficient)
	return &Mash{
		Units: "L",
		k: k,
		tankArea: tankArea,
		valveArea: valveArea,
	}, nil
}

func (m *Mash) Start(in <-chan gogadgets.Message, out chan<- gogadgets.Value) {
	m.out = out
	m.stop = make(chan bool)
	for {
		msg := <-in
		m.readMessage(msg)
	}
}

func (m *Mash) GetValue() *gogadgets.Value {
	return &gogadgets.Value{
		Value: m.Volume,
		Units: m.Units,
	}
}

func (m *Mash) sendCurrentVolume(startVolume float64, duration time.Duration) {
	m.Volume = m.previousVolume + m.getVolume(startVolume, duration.Seconds())
	m.out<- gogadgets.Value{
		Value: m.Volume,
		Units: m.Units,
	}
}

func (m *Mash) readMessage(msg gogadgets.Message) {
	if msg.Sender == "mash tun valve" {
		if msg.Value.Value == true {
			m.valveStatus = true
			go m.monitor(m.stop)
		} else if msg.Value.Value == false && m.valveStatus{
			m.valveStatus = false
			m.previousVolume = m.Volume
			m. stop<- true
		}
	} else if msg.Sender == "hlt volume" {
		m.hltVolume = msg.Value.Value.(float64)
	}
}

func (m *Mash) monitor(stop <-chan bool) {
	startTime := time.Now()
	interval := time.Duration(100 * time.Millisecond)
	startVolume := m.hltVolume * 1000.0
	var d time.Duration
	for  {
		select {
		case s := <-stop:
			fmt.Println(s)
			break
		case <-time.After(interval):
			if m.valveStatus {
				d = time.Since(startTime)
				m.sendCurrentVolume(startVolume, d)
			} else {
				interval = time.Duration(100 * time.Second)
			}
		}
	}
	fmt.Println("monitor exit")
}

func (m *Mash) getVolume(startVolume, elapsedTime float64) float64 {	
	height := m.getHeight(startVolume)
	dh := math.Abs(math.Pow((elapsedTime / m.k), 2) - (2 * (elapsedTime / m.k) * math.Pow(height, 0.5)))
	return (m.tankArea * dh) / 1000.0
}

func (m *Mash) getHeight(volume float64) float64 {
	return volume / m.tankArea
}
