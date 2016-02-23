package brewery

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cswank/gogadgets"
)

const (
	TOGALLONS = 0.264172
)

var (
	all = []string{"hlt", "tun", "boiler", "carboy"}
)

type volumeGetter func() float64

//BrewVolume keeps track of the volume in all three tanks
//(hlt, mash tun, and boiler) in by brew system.  It knows
//how much water is in the hlt when the float switch (which
//is watched by BrewVolume.poller) it triggered.  It waits
//for messages from one of the three solenoid valves.  If
//the valve that dumps water from the hlt to the mash tun
//it activated then it uses and equation of water flowing
//through an orifice to determine the volume at a given time.
//This allows the system to fill the mash tun with a specific
//amount of water.  Once the mash tun has water in it, then
//if the valve that controls the water flow the boiler is
//activated then a different strategy is used.  Since the
//mash tun is full of grains the water flow rate is unpredictable.
//This system just waits for a safe amount of time (BrewVolume.boilerFillTime)
//to say all the water that used to be in the mash tun is
//now in the boiler.  The boiler drains into the fermenter
//using a pump.  BrewVolume waits for a message indicating
//the pump is turned off then it puts all of the boiler volume
//in the fermenter.
type BrewVolume struct {
	hltVolume      float64
	tunVolume      float64
	boilerVolume   float64
	carboyVolume   float64
	hltCapacity    float64
	hltFull        chan bool
	mashArea       float64
	mashStop       chan bool
	boilStop       chan bool
	poller         gogadgets.Poller
	out            chan<- gogadgets.Message
	lock           sync.Mutex
	k              float64
	boilerFillTime time.Duration
	listening      bool
	volumes        map[string]volumeGetter
}

type BrewConfig struct {
	MashRadius      float64
	MashValveRadius float64
	Coefficient     float64
	HLTCapacity     float64
	Poller          gogadgets.Poller
	BoilerFillTime  int //time to drain the mash in seconds
	FloatSwitchPin  string
	FloatSwitchPort string
}

func NewBrewVolume(cfg *BrewConfig) (*BrewVolume, error) {
	k, mashArea := getK(cfg)
	b := &BrewVolume{
		hltFull:        make(chan bool),
		mashStop:       make(chan bool),
		boilStop:       make(chan bool),
		mashArea:       mashArea,
		hltCapacity:    cfg.HLTCapacity,
		k:              k,
		poller:         cfg.Poller,
		boilerFillTime: time.Duration(cfg.BoilerFillTime) * time.Second,
	}

	b.volumes = map[string]volumeGetter{
		"hlt":    b.getHLTVolume,
		"tun":    b.getTunVolume,
		"boiler": b.getBoilerVolume,
		"carboy": b.getCarboyVolume,
	}

	return b, nil
}

func getK(cfg *BrewConfig) (float64, float64) {
	mashArea := math.Pi * math.Pow(cfg.MashRadius, 2)
	valveArea := math.Pi * math.Pow(cfg.MashValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	return (mashArea * x) / (valveArea * cfg.Coefficient), mashArea
}

func (b *BrewVolume) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	b.out = out
	b.sendUpdates(all)

	for {
		select {
		case msg := <-input:
			b.readMessage(msg)
		case f := <-b.hltFull:
			b.updateHLTVolume(f)
		}
	}
}

func (b *BrewVolume) GetDirection() string {
	return "input"
}

func (b *BrewVolume) GetUID() string {
	return "brew volume"
}

func (b *BrewVolume) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		b.sendUpdates(all)
	} else if msg.Type == "update" && msg.Sender == "hlt valve" && msg.Value.Value == true {
		go b.waitForFloatSwitch()
	} else if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == true {
		go b.updateMashTunVolume()
	} else if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == false {
		b.mashStop <- true
	} else if msg.Type == "update" && msg.Sender == "boiler valve" && msg.Value.Value == true {
		go b.updateBoilerVolume()
	} else if msg.Type == "update" && msg.Sender == "boiler valve" && msg.Value.Value == false && b.listening {
		b.boilStop <- true
	} else if msg.Type == "update" && msg.Sender == "carboy pump" && msg.Value.Value == false {
		b.updateCarboyVolume()
	}
}

func (b *BrewVolume) updateMashTunVolume() {
	v := map[string]float64{
		"tun volume": b.tunVolume,
		"hlt volume": b.hltVolume,
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

//The rate that water flows from the mash into the boiler isn't
//known, so wait for a safe amount of time and assume all water
//in the mash is now in the boiler.
func (b *BrewVolume) updateBoilerVolume() {
	b.listening = true
	for b.listening {
		select {
		case <-b.boilStop:
			b.listening = false
		case <-time.After(b.boilerFillTime): //todo make this time part of the config
			b.boilerVolume = b.tunVolume
			b.tunVolume = 0.0
			b.sendUpdates([]string{"tun", "boiler"})
			b.listening = false
		}
	}
}

func (b *BrewVolume) updateCarboyVolume() {
	b.carboyVolume = b.boilerVolume
	b.boilerVolume = 0.0
	b.sendUpdates([]string{"boiler", "carboy"})
}

func (b *BrewVolume) sendUpdates(tanks []string) {
	b.lock.Lock()
	for _, t := range tanks {
		b.sendUpdate(t, "volume", b.volumes[t]())
	}
	b.lock.Unlock()
}

func (b *BrewVolume) sendUpdate(location, name string, value float64) {
	b.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    fmt.Sprintf("%s %s", location, name),
		Location:  location,
		Name:      name,
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: value * TOGALLONS,
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}

func (b *BrewVolume) getNewVolumes(start time.Time, startingVolumes map[string]float64) {
	d := time.Now().Sub(start)
	b.lock.Lock()
	defer b.lock.Unlock()
	b.calculateVolumes(d, startingVolumes)
	b.sendUpdate("hlt", "volume", b.hltVolume)
	b.sendUpdate("tun", "volume", b.tunVolume)
}

func (b *BrewVolume) calculateVolumes(duration time.Duration, startVolumes map[string]float64) {
	v := b.newVolume(startVolumes["hlt volume"], duration.Seconds())
	b.hltVolume = startVolumes["hlt volume"] - v
	b.tunVolume = startVolumes["tun volume"] + v
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
	b.hltVolume = b.hltCapacity
	b.sendUpdate("hlt", "volume", b.hltVolume)
}

func (b *BrewVolume) getHLTVolume() float64 {
	return b.hltVolume
}

func (b *BrewVolume) getTunVolume() float64 {
	return b.tunVolume
}

func (b *BrewVolume) getBoilerVolume() float64 {
	return b.boilerVolume
}

func (b *BrewVolume) getCarboyVolume() float64 {
	return b.carboyVolume
}
