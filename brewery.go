package brewery

import (
	"log"
	"time"

	"github.com/cswank/gogadgets"
)

var (
	vol *volumeManager
)

type BrewConfig struct {
	HLTRadius       float64
	TunValveRadius  float64
	HLTCoefficient  float64
	HLTCapacity     float64
	Poller          gogadgets.Poller
	BoilerFillTime  int //time to drain the mash in seconds
	FloatSwitchPin  string
	FloatSwitchPort string
}

func NewBrewery(cfg *BrewConfig) (*HLT, *Tun, *Boiler, *Carboy) {
	hlt, err := getHLT(cfg)
	if err != nil {
		log.Fatal(err)
	}

	tun, err := getTun(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fillTime := time.Duration(cfg.BoilerFillTime) * time.Second

	boiler, err := NewBoiler(fillTime)
	if err != nil {
		log.Fatal(err)
	}

	carboy, err := NewCarboy()
	if err != nil {
		log.Fatal(err)
	}
	return hlt, tun, boiler, carboy
}

func getTun(cfg *BrewConfig) (*Tun, error) {
	tunCfg := &TunConfig{
		Coefficient:    cfg.HLTCoefficient,
		HLTRadius:      cfg.HLTRadius,
		TunValveRadius: cfg.TunValveRadius,
	}
	return NewTun(tunCfg)
}

func getHLT(cfg *BrewConfig) (*HLT, error) {

	hltCfg := &HLTConfig{
		HLTCapacity: cfg.HLTCapacity,
		Poller:      getPoller(cfg),
	}
	return NewHLT(hltCfg)
}

//gpio for the float switch at the top of my hlt.  When
//it is triggered I know how much water is in the container.
func getPoller(cfg *BrewConfig) gogadgets.Poller {
	pin := &gogadgets.Pin{
		Port:      cfg.FloatSwitchPort,
		Pin:       cfg.FloatSwitchPin,
		Direction: "in",
		Edge:      "rising",
	}

	g, err := gogadgets.NewGPIO(pin)
	if err != nil {
		panic(err)
	}
	return g
}
