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

type Tun struct {
	volumes   map[string]float64
	tunArea   float64
	stop      chan bool
	out       chan<- gogadgets.Message
	k         float64
	listening bool
	lock      sync.Mutex
	after     Afterer
	filling   bool
	timer     Timer
}

type TunConfig struct {
	TunRadius      float64
	TunValveRadius float64
	Coefficient    float64
	HLTCapacity    float64
	After          Afterer
	Timer          Timer
}

func getK(c *TunConfig) (float64, float64) {
	hltArea := math.Pi * math.Pow(c.TunRadius, 2)
	valveArea := math.Pi * math.Pow(c.TunValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	return (hltArea * x) / (valveArea * c.Coefficient), hltArea
}

func NewTun(cfg *TunConfig) (*Tun, error) {
	k, tunArea := getK(cfg)
	if cfg.After == nil {
		cfg.After = time.After
	}
	if cfg.Timer == nil {
		cfg.Timer = &timer{}
	}
	return &Tun{
		volumes: map[string]float64{
			"hlt":    0.0,
			"tun":    0.0,
			"boiler": 0.0,
		},
		stop:    make(chan bool),
		tunArea: tunArea,
		k:       k,
		after:   cfg.After,
		timer:   cfg.Timer,
	}, nil
}

func (t *Tun) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	t.out = out
	t.sendUpdate()

	for {
		msg := <-input
		t.readMessage(msg)
	}
}

func (t *Tun) GetDirection() string {
	return "input"
}

func (t *Tun) GetUID() string {
	return "tun volume"
}

func (t *Tun) readMessage(msg gogadgets.Message) {
	if msg.Type == "command" && msg.Body == "update" {
		t.sendUpdate()
	} else if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == true {
		go t.fill()
	} else if msg.Type == "update" && msg.Sender == "tun valve" && msg.Value.Value == false && t.filling {
		t.stop <- true
	} else if msg.Type == "update" && msg.Sender == "hlt volume" {
		t.lock.Lock()
		t.volumes["hlt"] = msg.Value.Value.(float64) * (1.0 / TOGALLONS)
		t.lock.Unlock()
	} else if msg.Type == "update" && msg.Sender == "boiler volume" {
		t.lock.Lock()
		t.volumes["tun"] = 0.0
		t.lock.Unlock()
		t.sendUpdate()
	}
}

func (t *Tun) fill() {
	t.filling = true
	v := map[string]float64{
		"tun": t.volumes["tun"],
		"hlt": t.volumes["hlt"],
	}
	t.timer.Start()
	for {
		select {
		case <-t.stop:
			t.filling = false
			t.getNewVolumes(v)
			return
		case <-t.after(100 * time.Millisecond):
			t.getNewVolumes(v)
		}
	}
}

func (t *Tun) sendUpdate() {
	t.out <- gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "tun volume",
		Location:  "tun",
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: t.volumes["tun"] * TOGALLONS,
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
}

func (t *Tun) getNewVolumes(startVolumes map[string]float64) {
	t.lock.Lock()
	v := t.newVolume(startVolumes["hlt"], t.timer.Since().Seconds())
	t.volumes["tun"] = startVolumes["tun"] + v
	t.lock.Unlock()
	t.sendUpdate()
}

func (t *Tun) newVolume(startVolume, elapsedTime float64) float64 {
	height := startVolume / t.tunArea
	dh := math.Abs(math.Pow((elapsedTime/t.k), 2) - (2 * (elapsedTime / t.k) * math.Pow(height, 0.5)))
	if math.IsNaN(dh) {
		dh = 0.0
	}
	return (t.tunArea * dh) / 1000.0
}
