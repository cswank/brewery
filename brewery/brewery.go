package main

import (
	"flag"
	"io/ioutil"

	"bitbucket.org/cswank/brewery/brewgadgets"
	"bitbucket.org/cswank/gogadgets"
	"bitbucket.org/cswank/gogadgets/utils"
)

var (
	cfg = flag.String("c", "", "Path to the config json file")
)

func main() {
	flag.Parse()
	if !utils.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}
	gpio, err := gogadgets.NewGPIO(&gogadgets.Pin{Port: "8", Pin: "9", Direction: "in", Edge: "rising"})
	if err != nil {
		panic(err)
	}
	mashConfig := &brewgadgets.MashConfig{
		TankRadius:  7.5 * 2.54,
		ValveRadius: 0.1875 * 2.54,
		Coefficient: 0.4,
	}
	poller := gpio.(gogadgets.Poller)
	a, err := getApp(*cfg, mashConfig, poller, nil)
	if err != nil {
		panic(err)
	}
	a.Start()
}

func getApp(cfg interface{}, mashConfig *brewgadgets.MashConfig, poller gogadgets.Poller, boilerConfig *brewgadgets.BoilerConfig) (*gogadgets.App, error) {
	a := gogadgets.NewApp(cfg)

	mashVolume, _ := brewgadgets.NewMash(mashConfig)

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

	hltVolume := &brewgadgets.HLT{
		GPIO:  poller,
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

	boilerVolume, err := brewgadgets.NewBoiler(boilerConfig)
	if err != nil {
		return nil, err
	}
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
	return a, nil
}
