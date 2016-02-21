package main

import (
	"flag"
	"io/ioutil"

	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"
	"github.com/cswank/gogadgets/utils"
)

var (
	cfg = flag.String("c", "", "Path to the config json file")
)

func main() {
	flag.Parse()
	if !utils.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}

	pin := &gogadgets.Pin{Port: "8", Pin: "9", Direction: "in", Edge: "rising"}
	gpio, err := gogadgets.NewGPIO(pin)
	if err != nil {
		panic(err)
	}

	brewCfg := &brewery.BrewConfig{
		MashRadius:      7.5 * 2.54,
		MashValveRadius: 0.1875 * 2.54,
		HLTCapacity:     26.5,
		Coefficient:     0.4,
		BoilerFillTime:  60 * 5,
		Poller:          gpio.(gogadgets.Poller),
	}

	a, err := getApp(*cfg, brewCfg)
	if err != nil {
		panic(err)
	}
	a.Start()
}

func getApp(cfg interface{}, brewCfg *brewery.BrewConfig) (*gogadgets.App, error) {
	a := gogadgets.NewApp(cfg)

	vol, err := brewery.NewBrewVolume(brewCfg)
	if err != nil {
		return nil, err
	}

	a.AddGadget(vol)

	return a, nil
}
