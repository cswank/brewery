package brewery

import (
	"log"
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

type fit struct {
	a, b, c float64
}

type volumeManager struct {
	lock       sync.Mutex
	updateLock sync.Mutex
	volumes    map[string]float64
	updates    map[string]func(float64)
	stop       chan bool

	fit       fit
	listening bool

	hltCapacity float64

	after             Afterer
	fillingTun        bool
	fillingBoiler     bool
	boilerFillTime    time.Duration
	stopFillingBoiler chan bool
	stopFillingTun    chan bool
	timer             Timer
	poller            gogadgets.Poller
}

func newVolumeManager(cfg *Config, opts ...func(*volumeManager)) (*volumeManager, error) {
	v := &volumeManager{
		volumes: map[string]float64{
			"hlt": 0.0,
			"tun": 0.0,
		},
		updates:           map[string]func(float64){},
		stop:              make(chan bool),
		fit:               fit{a: cfg.A, b: cfg.B, c: cfg.C},
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
	v.updateLock.Lock()
	x := v.volumes[k]
	v.updateLock.Unlock()
	return x * mlToGallons
}

func (v *volumeManager) set(k string, val float64) {
	v.updateLock.Lock()
	v.volumes[k] = val * gallonsToML
	v.updateLock.Unlock()
}

func (v *volumeManager) readMessage(msg gogadgets.Message) {
	if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == true {
		go v.fillTun()
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
		v.updates["carboy"](v.get("carboy"))
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
			v.updates["boiler"](v.get("boiler"))
			v.updates["tun"](0.0)
		}
	}
}

func (v *volumeManager) waitForFloatSwitch() {
	err := v.poller.Wait()
	if err != nil {
		log.Println("gpio Wait() error", err)
		return
	}

	v.lock.Lock()
	v.volumes["hlt"] = v.hltCapacity * gallonsToML
	v.lock.Unlock()
	v.updates["hlt"](v.hltCapacity)
}

func (v *volumeManager) register(k string, f func(float64)) {
	v.updates[k] = f
}

func (v *volumeManager) fillTun() {
	v.fillingTun = true
	v.lock.Lock()
	m := map[string]float64{
		"tun": v.volumes["tun"],
		"hlt": v.volumes["hlt"],
	}
	v.lock.Unlock()
	v.timer.Start()
	for {
		select {
		case <-v.stopFillingTun:
			v.fillingTun = false
			v.getNewVolume(m)
			return
		case <-v.after(time.Second):
			v.getNewVolume(m)
		}
	}
}

func (v *volumeManager) getNewVolume(startVolumes map[string]float64) {
	vol := v.newVolume(v.timer.Since().Seconds())
	v.lock.Lock()
	v.volumes["hlt"] = startVolumes["hlt"] - vol
	v.volumes["tun"] = startVolumes["tun"] + vol
	v.lock.Unlock()
	v.updates["hlt"](v.get("hlt"))
	v.updates["tun"](v.get("tun"))
}

func (v *volumeManager) newVolume(elapsedTime float64) float64 {
	return v.fit.a + (v.fit.b * elapsedTime) + (v.fit.c * elapsedTime * elapsedTime)
}

//gpio for the float switch at the top of my hlt.  When
//it is triggered I know how much water is in the container.
func newPoller(cfg *Config) (gogadgets.Poller, error) {
	pin := &gogadgets.Pin{
		Pin:       cfg.FloatSwitchPin,
		Platform:  "rpi",
		Direction: "in",
		Edge:      "rising",
	}

	g, err := gogadgets.NewGPIO(pin)
	if err != nil {
		return nil, err
	}
	return g.(*gogadgets.GPIO), nil
}
