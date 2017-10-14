package brewery

import (
	"math"
	"sync"
	"time"

	"github.com/cswank/gogadgets"
)

const (
	//volume is tracked internally in mL.
	gallonsToML = 3785.41
	mlToGallons = 1.0 / gallonsToML
)

type Afterer func(d time.Duration) <-chan time.Time

type Timer interface {
	Start()
	Since() time.Duration
}

type timer struct {
	start time.Time
}

func (t *timer) Start() {
	t.start = time.Now()
}

func (t *timer) Since() time.Duration {
	return time.Now().Sub(t.start)
}

type volumeManager struct {
	lock      sync.Mutex
	volumes   map[string]float64
	updates   map[string]func(float64)
	hltArea   float64
	stop      chan bool
	k         float64
	listening bool

	hltCapacity float64

	after             Afterer
	fillingTun        bool
	fillingBoiler     bool
	boilerFillTime    time.Duration
	stopFillingBoiler chan bool
	stopFillingTun    chan bool
	timer             Timer
	target            float64
	poller            gogadgets.Poller
}

func newVolumeManager(cfg *Config, opts ...func(*volumeManager)) (*volumeManager, error) {
	k, hltArea := cfg.getK()

	v := &volumeManager{
		volumes: map[string]float64{
			"hlt": 0.0,
			"tun": 0.0,
		},
		updates:           map[string]func(float64){},
		stop:              make(chan bool),
		hltArea:           hltArea,
		k:                 k,
		boilerFillTime:    time.Duration(cfg.BoilerFillTime) * time.Second,
		stopFillingBoiler: make(chan bool),
		stopFillingTun:    make(chan bool),
		hltCapacity:       cfg.HLTCapacity,
	}

	for _, opt := range opts {
		opt(v)

	}

	if v.after == nil {
		v.after = time.After
	}

	if v.timer == nil {
		v.timer = &timer{}
	}

	if v.poller == nil {
		var err error
		v.poller, err = newPoller(cfg)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (v *volumeManager) get(k string) float64 {
	v.lock.Lock()
	x := v.volumes[k]
	v.lock.Unlock()
	return x * mlToGallons
}

func (v *volumeManager) set(k string, val float64) {
	v.lock.Lock()
	v.volumes[k] = val * gallonsToML
	v.lock.Unlock()
}

func (v *volumeManager) readMessage(msg gogadgets.Message) {
	if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == true {
		go v.fillTun(msg.TargetValue)
	} else if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == false && v.fillingTun {
		v.stopFillingTun <- true
	} else if msg.Type == "update" && msg.Sender == "boiler valve" && msg.Value.Value == true {
		go v.fillBoiler()
	} else if msg.Type == "update" && msg.Sender == "boiler valve" && msg.Value.Value == false && v.fillingBoiler {
		v.stopFillingBoiler <- true
	} else if msg.Type == "update" && msg.Sender == "hlt valve" && msg.Value.Value == true {
		go v.waitForFloatSwitch()
	} else if msg.Type == "update" && msg.Sender == "carboy pump" && msg.Value.Value == false {
		b := v.volumes["boiler"]
		v.volumes["carboy"] = b
		v.volumes["boiler"] = 0.0
		v.updates["carboy"](b)
		v.updates["boiler"](0.0)
	}
}

//The rate that water flows from the mash into the boiler isn't
//known, so wait for a safe amount of time and assume all water
//in the mash is now in the boiler.
func (v *volumeManager) fillBoiler() {
	v.fillingBoiler = true
	for v.fillingBoiler {
		select {
		case <-v.stopFillingBoiler:
			v.fillingBoiler = false
		case <-time.After(v.boilerFillTime):
			v.fillingBoiler = false
			v.lock.Lock()
			t := v.volumes["tun"]
			v.volumes["boiler"] = t
			v.volumes["tun"] = 0.0
			v.lock.Unlock()
			v.updates["boiler"](t)
			v.updates["tun"](0.0)

		}
	}
}

func (v *volumeManager) waitForFloatSwitch() {
	b, _ := v.poller.Wait()
	if !b {
		return
	}

	v.lock.Lock()
	v.volumes["hlt"] = v.hltCapacity
	v.lock.Unlock()
	v.updates["hlt"](v.hltCapacity)
}

func (v *volumeManager) register(k string, f func(float64)) {
	v.updates[k] = f
}

func (v *volumeManager) fillTun(val *gogadgets.Value) {
	v.fillingTun = true
	if val != nil {
		v.target = val.Value.(float64)
	}
	v.lock.Lock()
	m := map[string]float64{
		"tun": v.volumes["tun"],
		"hlt": v.volumes["hlt"],
	}
	v.lock.Unlock()
	v.timer.Start()
	var i int
	for {
		select {
		case <-v.stopFillingTun:
			v.fillingTun = false
			v.target = 0.0
			v.getNewVolume(m, 0)
			return
		case <-v.after(time.Second):
			v.getNewVolume(m, i)
			i++
		}
	}
}

func (v *volumeManager) getNewVolume(startVolumes map[string]float64, i int) {
	vol := v.newVolume(startVolumes["hlt"], v.timer.Since().Seconds())
	t := startVolumes["tun"] + vol
	h := startVolumes["hlt"] - vol
	v.lock.Lock()
	v.volumes["hlt"] = h
	v.volumes["tun"] = t
	v.lock.Unlock()
	v.updates["hlt"](h)
	v.updates["tun"](t)
}

func (v *volumeManager) getTime(volume, startVolume float64) time.Duration {
	h := startVolume / v.hltArea
	dh := volume / v.hltArea
	k2 := v.k * v.k
	x := (math.Pow(h, 0.5) * v.k) - math.Pow((dh*k2)+h*k2, 0.5)
	return time.Duration(x) * time.Second
}

func (v *volumeManager) newVolume(startVolume, elapsedTime float64) float64 {
	height := startVolume / v.hltArea
	dh := math.Abs(math.Pow((elapsedTime/v.k), 2) - (2 * (elapsedTime / v.k) * math.Pow(height, 0.5)))
	if math.IsNaN(dh) {
		dh = 0.0
	}
	return v.hltArea * dh
}

//gpio for the float switch at the top of my hlt.  When
//it is triggered I know how much water is in the container.
func newPoller(cfg *Config) (gogadgets.Poller, error) {
	pin := &gogadgets.Pin{
		Port:      cfg.FloatSwitchPort,
		Pin:       cfg.FloatSwitchPin,
		Direction: "in",
		Edge:      "rising",
	}

	return gogadgets.NewGPIO(pin)
}
