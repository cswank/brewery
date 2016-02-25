package brewery

import (
	"math"
	"time"

	"github.com/cswank/gogadgets"
)

type Boiler struct {
	volume   float64
	fillTime time.Duration
}

type BoilerConfig struct {
	FillTime int //time to drain the mash in seconds
}

func NewBoiler(fillTime int) (*BrewVolume, error) {
	return &Boiler{
		fillTime: time.Duration(fillTime) * time.Second,
	}, nil
}

func (b *Boiler) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
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
