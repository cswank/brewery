package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"
	"github.com/cswank/gogadgets/utils"
	"github.com/kelseyhightower/envconfig"
)

var (
	cfg = flag.String("c", "", "Path to the config json file")
)

func main() {
	flag.Parse()
	var brewCfg brewery.BrewConfig
	if err := envconfig.Process("brewery", &brewCfg); err != nil {
		log.Fatal(err)
	}

	checkW1()

	brewCfg.Poller = getPoller(brewCfg)

	a, err := getApp(*cfg, &brewCfg)
	if err != nil {
		panic(err)
	}
	a.Start()
}

//gpio for the float switch at the top of my hlt.  When
//it is triggered I know how much water is in the container.
func getPoller(cfg brewery.BrewConfig) gogadgets.Poller {
	pin := &gogadgets.Pin{
		Port:      cfg.FloatSwitchPort,
		Pin:       cfg.FloatSwitchPin,
		Direction: "in",
		Edge:      "rising",
	}

	gpio, err := gogadgets.NewGPIO(pin)
	if err != nil {
		panic(err)
	}
	return gpio.(gogadgets.Poller)
}

//I am too lazy to load the BB-W1 device tree overlay the right way, which would
//be to add it to uEnv.txt (which is a config file for u-boot).
func checkW1() {
	if !utils.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}
}

func getApp(cfg interface{}, brewCfg *brewery.BrewConfig) (*gogadgets.App, error) {
	a := gogadgets.NewApp(cfg)

	// vol, err := brewery.NewBrewVolume(brewCfg)
	// if err != nil {
	// 	return nil, err
	// }

	//a.AddGadget(vol)
	return a, nil
}
