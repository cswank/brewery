package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/cswank/gogadgets"
)

var (
	configPath  = flag.String("c", "", "Path to the config json file")
	hltVolume   = flag.Float64("V", 0.0, "HLT Volume")
	mashRadius  = flag.Float64("r", 0.0, "mash tun radius")
	valveRadius = flag.Float64("v", 0.0, "mash tun valve radius")
)

func main() {
	flag.Parse()
	if !gogadgets.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}
	b, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	cfg := &gogadgets.Config{}
	err = json.Unmarshal(b, cfg)
	for _, config := range cfg.Gadgets {
		if config.Location == "mash tun" && config.Name == "valve" {
			gpio, err := gogadgets.NewGPIO(&config.Pin)
			if err == nil {
				cfg := &brewery.MashConfig{
					TankRadius:  *mashRadius,
					ValveRadius: *valveRadius,
				}
				dev, mashErr := brewgadgets.NewMash(cfg)
				mash := dev.(*brewgadgets.Mash)
				mash.HLTVolume = *hltVolume
				if mashErr == nil {
					brewgadgets.Calibrate(mash, gpio)
				} else {
					fmt.Println(mashErr)
				}
			} else {
				fmt.Println(err)
			}
		}
	}
}
