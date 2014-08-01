package main

import (
	"bitbucket.org/cswank/gogadgets"
	"bitbucket.org/cswank/gogadgets/utils"
	"bitbucket.org/cswank/brewery/brewgadgets"
	"flag"
	"io/ioutil"
)

var (
	cfg = flag.String("c", "", "Path to the config json file")
)

func main() {
	flag.Parse()
	if !utils.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}
	
	a := gogadgets.NewApp(*cfg)

	config := &brewgadgets.MashConfig{
		TankRadius:  7.5 * 2.54,
		ValveRadius: 0.1875 * 2.54,
		Coefficient: 0.4,
	}
	mashVolume, _ := brewgadgets.NewMash(config)

	mash := &gogadgets.Gadget{
		Location:   "tun",
		Name:       "volume",
		Input:      mashVolume,
		Direction:  "input",
		OnCommand:  "n/a",
		OffCommand: "n/a",
		UID:        "tun volume",
	}

	a.AddGadget(mash)
	poller, err := gogadgets.NewGPIO(&gogadgets.Pin{Port: "8", Pin: "9", Direction: "in", Edge: "rising"})
	if err != nil {
		panic(err)
	}

	hltVolume := &brewgadgets.HLT{
		GPIO:  poller.(gogadgets.Poller),
		Value: 7.0,
		Units: "gallons",
	}
	hlt := &gogadgets.Gadget{
		Location:   "hlt",
		Name:       "volume",
		Input:      hltVolume,
		Direction:  "input",
		OnCommand:  "n/a",
		OffCommand: "n/a",
		UID:        "hlt volume",
	}
	a.AddGadget(hlt)

	boilerVolume, _ := brewgadgets.NewBoiler()
	boiler := &gogadgets.Gadget{
		Location:   "boiler",
		Name:       "volume",
		Input:      boilerVolume,
		Direction:  "input",
		OnCommand:  "n/a",
		OffCommand: "n/a",
		UID:        "boiler volume",
	}
	a.AddGadget(boiler)
	a.Start()
}
