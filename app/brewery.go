package main


import (
	"bitbucket.com/cswank/gogadgets"
	"bitbucket.com/cswank/brewery/gogadgets"
	"encoding/json"
	"io/ioutil"
	"flag"
)

func main() {
	flag.Parse()
	if !utils.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}
	b, err := ioutil.ReadFile(*configFlag)
	if err != nil {
		panic(err)
	}
	cfg := &gogadgets.Config{}
	err = json.Unmarshal(b, cfg)
	a := gogadgets.NewApp(cfg)

	config = &MashConfig{
		TankRadius: 7.5 * 2.54,
		ValveRadius: 0.1875 * 2.54,
		Coefficient: 0.43244,
	}
	mash, _ := NewMash(config)
	a.AddGadget(mash)
	poller, err := NewGPIO(&Pin{Port:"8", Pin:"9", Direction:"in", Edge:"rising"})
	if err != nil {
		panic(err)
	}
	hlt := &HLT{
		GPIO: poller,
		volume: 26.5,
		units: "liters",
	}
	a.AddGadget(hlt)
}
