package main

import (
	"bitbucket.org/cswank/gogadgets"
	"bitbucket.org/cswank/gogadgets/models"
	"bitbucket.org/cswank/gogadgets/input"
	"bitbucket.org/cswank/gogadgets/output"
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
		Location:   "mash tun",
		Name:       "volume",
		Input:      mashVolume,
		Direction:  "input",
		OnCommand:  "n/a",
		OffCommand: "n/a",
		UID:        "mash tun volume",
	}

	a.AddGadget(mash)
	poller, err := output.NewGPIO(&models.Pin{Port: "8", Pin: "9", Direction: "in", Edge: "rising"})
	if err != nil {
		panic(err)
	}

	hltVolume := &brewgadgets.HLT{
		GPIO:  poller.(input.Poller),
		Value: 26.5,
		Units: "liters",
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
