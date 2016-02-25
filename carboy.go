package brewery

import (
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
	volumes        map[string]float64
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
		volumes: map[string]float64{
			"hlt":    0.0,
			"tun":    0.0,
			"boiler": 0.0,
			"carboy": 0.0,
		},
		hltFull:        make(chan bool),
		mashStop:       make(chan bool),
		boilStop:       make(chan bool),
		mashArea:       mashArea,
		hltCapacity:    cfg.HLTCapacity,
		k:              k,
		poller:         cfg.Poller,
		boilerFillTime: time.Duration(cfg.BoilerFillTime) * time.Second,
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
		"tun": b.volumes["tun"],
		"hlt": b.volumes["hlt"],
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
			b.volumes["boiler"] = b.volumes["tun"]
			b.volumes["tun"] = 0.0
			b.sendUpdates([]string{"tun", "boiler"})
			b.listening = false
		}
	}
}

func (b *BrewVolume) updateCarboyVolume() {
	b.volumes["carboy"] = b.volumes["boiler"]
	b.volumes["boiler"] = 0.0
	b.sendUpdates([]string{"boiler", "carboy"})
}

func (b *BrewVolume) sendUpdates(tanks []string) {
	b.lock.Lock()
	for _, t := range tanks {
		b.sendUpdate(t)
	}
	b.lock.Unlock()
}

func (b *BrewVolume) sendUpdate(location string) {
	b.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "brew volume",
		Location:  location,
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: b.volumes[location] * TOGALLONS,
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
	b.sendUpdate("hlt")
	b.sendUpdate("tun")
}

func (b *BrewVolume) calculateVolumes(duration time.Duration, startVolumes map[string]float64) {
	v := b.newVolume(startVolumes["hlt"], duration.Seconds())
	b.volumes["hlt"] = startVolumes["hlt"] - v
	b.volumes["tun"] = startVolumes["tun"] + v
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
	b.volumes["hlt"] = b.hltCapacity
	b.sendUpdate("hlt")
}
