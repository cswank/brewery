package brewgadgets

import (
	"fmt"
	"time"
	"math"
	"bitbucket.com/cswank/gogadgets"
)

type Mash struct {
	gogadgets.InputDevice
	volume float64
	hltVolume float64
	Units string
	out chan<- gogadgets.Value
	k float64
	tankArea float64
	valveArea float64
	targetVolume float64
	valveStatus bool
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
	for {
		msg := <- in
		m.readMessage(msg)
	}
}

func (m *Mash) monitor(stop <-chan bool) {
	startVolume := m.hltVolume
	drainTime := time.Duration(m.getDrainTime(startVolume, m.targetVolume) * float64(time.Second))
	startTime := time.Now()
	interval := time.Duration(100 * time.Millisecond)
	for interval.Seconds() > 0.0 {
		select {
		case <-stop:
			return
		case <-time.After(interval):
			interval = m.sendCurrentVolume(startVolume, startTime, drainTime)
		}
	}
}

func (m *Mash) sendCurrentVolume(startingVolume float64, startTime time.Time, drainTime time.Duration) time.Duration {
	d := time.Since(startTime)
	m.volume = m.getVolume(startingVolume, d.Seconds())
	m.out<- gogadgets.Value{
		Value: m.volume,
		Units: m.Units,
	}
	remaining := drainTime - d
	if remaining.Seconds() < 0.1 {
		return remaining
	}
	return time.Duration(100 * time.Millisecond)
}

func (m *Mash) readMessage(msg gogadgets.Message) {
	stop := make(chan bool)
	if msg.Type == "update" && msg.Sender == "hlt valve" {
		if msg.Value.Value == true {
			m.valveStatus = true
			go m.monitor(stop)
		} else if msg.Value.Value == false && m.valveStatus{
			m.valveStatus = false
			stop<- true
		}
	} else if msg.Type == "update" && msg.Sender == "hlt volume" {
		fmt.Println("setting hlt volume", msg)
		m.hltVolume = msg.Value.Value.(float64)
	} else if msg.Type == "command" && msg.Body == "fill mash tun" {
		fmt.Println("setting target volume")
		//get the target volume, if present, so a message can be
		//send when the target volume is achieved.
	}
}

func (m *Mash) getDrainTime(startingVolume, volume float64) float64 {
	volume = 1000 * volume //convert to cubic centimeters
	startingVolume = startingVolume * 1000.0
	heightDiff := m.getHeight(volume)
	height := m.getHeight(startingVolume)
	h2 := height - heightDiff
	return (math.Pow(height, 0.5) - math.Pow(h2, 0.5)) * m.k
}

func (m *Mash) getVolume(startingVolume, elapsedTime float64) float64 {
	startingVolume = startingVolume * 1000.0 //convert to cubic centimeters
	height := m.getHeight(startingVolume)
	dh := math.Abs(math.Pow((elapsedTime / m.k), 2) - (2 * (elapsedTime / m.k) * math.Pow(height, 0.5)))
	return (m.tankArea * dh) / 1000.0
}

func (m *Mash) getHeight(volume float64) float64 {
	return volume / m.tankArea
}








    

