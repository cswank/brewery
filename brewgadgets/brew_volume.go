package brewgadgets

import (
	"math"
	"sync"
	"time"

	"bitbucket.org/cswank/gogadgets"
)

const (
	TOGALLONS = 0.264172
)

type BrewVolume struct {
	HLTCapacity  float64
	hltVolume    float64
	mashVolume   float64
	boilerVolume float64
	hltFull      chan bool
	mashArea     float64
	mashStop     chan bool
	poller       gogadgets.Poller
	out          chan<- gogadgets.Message
	lock         sync.Mutex
	k            float64
}

type BrewConfig struct {
	MashRadius      float64
	MashValveRadius float64
	Coefficient     float64
	HLTCapacity     float64
}

func NewBrewVolume(config *BrewConfig) (*BrewVolume, error) {
	k, mashArea := getK(config)
	return &BrewVolume{
		hltFull:     make(chan bool),
		mashStop:    make(chan bool),
		mashArea:    mashArea,
		HLTCapacity: config.HLTCapacity,
		k:           k,
	}, nil
}

func getK(config *BrewConfig) (float64, float64) {
	mashArea := math.Pi * math.Pow(config.MashRadius, 2)
	valveArea := math.Pi * math.Pow(config.MashValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	return (mashArea * x) / (valveArea * config.Coefficient), mashArea
}

func (b *BrewVolume) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	b.out = out
	for {
		select {
		case msg := <-input:
			b.readMessage(msg)
		case f := <-b.hltFull:
			b.updateHLTVolume(f)
		}
	}
}

func (b *BrewVolume) GetUID() string {
	return "brew volume"
}

func (b *BrewVolume) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		b.sendUpdates()
	} else if msg.Type == "update" && msg.Sender == "hlt valve" && msg.Value.Value == true {
		go b.waitForFloatSwitch()
	} else if msg.Type == "update" && msg.Sender == "mash tun valve" && msg.Value.Value == true {
		go b.updateMashTunVolume()
	} else if msg.Type == "update" && msg.Sender == "mash tun valve" && msg.Value.Value == false {
		b.mashStop <- true
	}
}

func (b *BrewVolume) updateMashTunVolume() {
	v := map[string]float64{
		"mash tun volume": b.mashVolume,
		"hlt volume":      b.hltVolume,
	}
	ts := time.Now()
	for {
		select {
		case <-b.mashStop:
			b.getNewVolumes(ts, v)
			return
		case <-time.After(100 * time.Millisecond):
			b.getNewVolumes(ts, v)
		}
	}
}

func (b *BrewVolume) sendUpdates() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.sendUpdate("hlt volume", b.hltVolume)
	b.sendUpdate("mash tun volume", b.mashVolume)
	b.sendUpdate("boiler volume", b.boilerVolume)
}

func (b *BrewVolume) sendUpdate(sender string, value float64) {
	b.out <- gogadgets.Message{
		Sender: sender,
		Type:   "update",
		Value: gogadgets.Value{
			Value: value * TOGALLONS,
			Units: "gallons",
		},
	}
}

func (b *BrewVolume) getNewVolumes(start time.Time, startingVolumes map[string]float64) {
	d := time.Now().Sub(start)
	b.lock.Lock()
	defer b.lock.Unlock()
	b.calculateVolumes(d, startingVolumes)
	b.sendUpdate("hlt volume", b.hltVolume)
	b.sendUpdate("mash tun volume", b.mashVolume)
}

func (b *BrewVolume) calculateVolumes(duration time.Duration, startVolumes map[string]float64) {
	v := b.newVolume(startVolumes["hlt volume"], duration.Seconds())
	b.hltVolume = startVolumes["hlt volume"] - v
	b.mashVolume = startVolumes["mash tun volume"] + v
}

func (b *BrewVolume) newVolume(startVolume, elapsedTime float64) float64 {
	height := startVolume / b.mashArea
	dh := math.Abs(math.Pow((elapsedTime/b.k), 2) - (2 * (elapsedTime / b.k) * math.Pow(height, 0.5)))
	if math.IsNaN(dh) {
		dh = 0.0
	}
	return (b.mashArea * dh) / 1000.0
}

func (b *BrewVolume) waitForFloatSwitch() {
	f, _ := b.poller.Wait()
	b.hltFull <- f
}

func (b *BrewVolume) updateHLTVolume(full bool) {
	if !full {
		return
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.hltVolume = b.HLTCapacity
	b.sendUpdate("hlt volume", b.hltVolume)
}
