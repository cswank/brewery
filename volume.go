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
	hltArea   float64
	stop      chan bool
	k         float64
	listening bool

	after   Afterer
	filling bool
	timer   Timer
	target  float64
}

type volumeConfig struct {
	HLTRadius      float64
	TunValveRadius float64
	Coefficient    float64
	After          Afterer
	Timer          Timer
}

func (v *volumeConfig) getK() (float64, float64) {
	hltArea := math.Pi * math.Pow(v.HLTRadius, 2)
	valveArea := math.Pi * math.Pow(v.TunValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	return (hltArea * x) / (valveArea * v.Coefficient), hltArea
}

func newVolumeManager(cfg *volumeConfig) (*volumeManager, error) {
	k, hltArea := cfg.getK()
	if cfg.After == nil {
		cfg.After = time.After
	}
	if cfg.Timer == nil {
		cfg.Timer = &timer{}
	}
	return &volumeManager{
		volumes: map[string]float64{
			"hlt": 0.0,
			"tun": 0.0,
		},
		stop:    make(chan bool),
		hltArea: hltArea,
		k:       k,
		after:   cfg.After,
		timer:   cfg.Timer,
	}, nil
}

func (v *volumeManager) get(k string) float64 {
	v.lock.Lock()
	x := v.volumes[k]
	v.lock.Unlock()
	return x * mlToGallons
}

func (v *volumeManager) readMessage(msg gogadgets.Message) {
	if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == true {
		go v.fill(msg.TargetValue)
	} else if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == false && v.filling {
		v.stop <- true
	} else if msg.Type == "update" && msg.Sender == "hlt volume" {
		v.lock.Lock()
		v.volumes["hlt"] = msg.Value.Value.(float64) * gallonsToML
		v.lock.Unlock()
	} else if msg.Type == "update" && msg.Sender == "boiler volume" {
		v.lock.Lock()
		vol := v.volumes["tun"]
		v.volumes["tun"] = 0.0
		v.volumes["boiler"] = vol
		v.lock.Unlock()
	}
}

func (v *volumeManager) fill(val *gogadgets.Value) {
	v.filling = true
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
		case <-v.stop:
			v.filling = false
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
	v.lock.Lock()
	v.volumes["tun"] = startVolumes["tun"] + vol
	v.volumes["hlt"] = startVolumes["hlt"] - vol
	v.lock.Unlock()
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
