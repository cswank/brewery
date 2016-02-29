package brewery

import (
	"math"
	"sync"
	"time"

	"github.com/cswank/gogadgets"
)

const (
	//volume is tracked internally in mL.
	GALLONS_TO_ML = 3785.41
	ML_TO_GALLONS = 1.0 / GALLONS_TO_ML
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
	hltArea   float64
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
	HLTRadius      float64
	TunValveRadius float64
	Coefficient    float64
	After          Afterer
	Timer          Timer
}

func getK(c *TunConfig) (float64, float64) {
	hltArea := math.Pi * math.Pow(c.HLTRadius, 2)
	valveArea := math.Pi * math.Pow(c.TunValveRadius, 2)
	g := 9.806 * 100.0 //centimeters
	x := math.Pow((2.0 / g), 0.5)
	return (hltArea * x) / (valveArea * c.Coefficient), hltArea
}

func NewTun(cfg *TunConfig) (*Tun, error) {
	k, hltArea := getK(cfg)
	if cfg.After == nil {
		cfg.After = time.After
	}
	if cfg.Timer == nil {
		cfg.Timer = &timer{}
	}
	return &Tun{
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
		//convert to mL
		t.volumes["hlt"] = msg.Value.Value.(float64) * GALLONS_TO_ML
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
	t.lock.Lock()
	v := map[string]float64{
		"tun": t.volumes["tun"],
		"hlt": t.volumes["hlt"],
	}
	t.lock.Unlock()
	t.timer.Start()
	var i int
	for {
		select {
		case <-t.stop:
			t.filling = false
			t.getNewVolume(v, 0)
			return
		case <-t.after(100 * time.Millisecond):
			t.getNewVolume(v, i)
			i++
		}
	}
}

func (t *Tun) sendUpdate() {
	msg := gogadgets.Message{
		UUID:      gogadgets.GetUUID(),
		Sender:    "tun volume",
		Location:  "tun",
		Name:      "volume",
		Type:      "update",
		Timestamp: time.Now().UTC(),
		Value: gogadgets.Value{
			Value: t.volumes["tun"] * ML_TO_GALLONS,
			Units: "gallons",
		},
		Info: gogadgets.Info{
			Direction: "input",
		},
	}
	t.out <- msg
}

func (t *Tun) getNewVolume(startVolumes map[string]float64, i int) {
	v := t.newVolume(startVolumes["hlt"], t.timer.Since().Seconds())
	t.lock.Lock()
	t.volumes["tun"] = startVolumes["tun"] + v
	if i%10 == 0 {
		t.sendUpdate()
	}
	t.lock.Unlock()
}

func (t *Tun) newVolume(startVolume, elapsedTime float64) float64 {
	height := startVolume / t.hltArea
	dh := math.Abs(math.Pow((elapsedTime/t.k), 2) - (2 * (elapsedTime / t.k) * math.Pow(height, 0.5)))
	if math.IsNaN(dh) {
		dh = 0.0
	}
	return (t.hltArea * dh) / 1000.0
}
