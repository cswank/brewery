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
	HLTVolume float64
	out chan<- gogadgets.Value
	k float64
	x float64
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
		x: x,
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
	m.Volume = m.previousVolume + m.GetVolume(startVolume, duration.Seconds())
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
		m.HLTVolume = msg.Value.Value.(float64)
	} else if msg.Sender == "boiler volume" {
		m.previousVolume = 0.0
		m.Volume = 0.0
		m.out<- *m.GetValue()
	}
}

func (m *Mash) monitor(stop <-chan bool) {
	startTime := time.Now()
	interval := time.Duration(100 * time.Millisecond)
	startVolume := m.HLTVolume * 1000.0
	var d time.Duration
	keepGoing := true
	for keepGoing {
		select {
		case <-stop:
			keepGoing = false
		case <-time.After(interval):
			if m.valveStatus {
				d = time.Since(startTime)
				m.sendCurrentVolume(startVolume, d)
			} else {
				
			}
		}
	}
	fmt.Println("exit monitor")
}

func (m *Mash) GetVolume(startVolume, elapsedTime float64) float64 {
	height := m.getHeight(startVolume)
	dh := math.Abs(math.Pow((elapsedTime / m.k), 2) - (2 * (elapsedTime / m.k) * math.Pow(height, 0.5)))
	if math.IsNaN(dh) {
		dh = 0.0
	}
	return (m.tankArea * dh) / 1000.0
}

func (m *Mash) GetDrainTime(startVolume, volume float64) float64 {
	volume = 1000 * volume //convert to cubic centimeters
	startVolume = startVolume * 1000.0
	heightDiff := m.getHeight(volume)
	height := m.getHeight(startVolume)
	h2 := height - heightDiff
	return (math.Pow(height, 0.5) - math.Pow(h2, 0.5)) * m.k
}

func (m *Mash) getHeight(volume float64) float64 {
	return volume / m.tankArea
}

func (m *Mash) GetCoefficient(startVolume, volume, drainTime float64) float64 {
	hi := m.getHeight(startVolume * 1000.0)
	dh := m.getHeight(volume * 1000.0)
	hf := hi - dh
	At := m.tankArea
	Av := m.valveArea
	t := drainTime
	return ((At * m.x) / Av) * (math.Pow(hi, 0.5) - math.Pow(hf, 0.5)) / t
}

func getLiter(mash *Mash, gpio gogadgets.OutputDevice) float64 {
	fmt.Println("Push enter to start")
	fmt.Scanf("x")
	gpio.On(nil)
	start := time.Now()
	fmt.Println("Push enter when 1 liters has dispensed")
	fmt.Scanf("x")
	gpio.Off()
	end := time.Now()
	duration := end.Sub(start)
	mash.HLTVolume -= 1.0
	return mash.GetCoefficient(mash.HLTVolume, 1.0, duration.Seconds())
}
	
func Calibrate(mash *Mash, mashValve gogadgets.OutputDevice) {
	coefficients := make([]float64, 5)
	for i := 0; i < 5; i++ {
		coefficients[i] = getLiter(mash, mashValve)
		fmt.Println(coefficients[i])
	}
	fmt.Println(coefficients)
}
	
	
	
