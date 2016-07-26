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
	cfg = flag.String("c", "", "Path to the gogadgets config json file")
)

func main() {
	flag.Parse()
	var brewCfg brewery.BrewConfig
	if err := envconfig.Process("brewery", &brewCfg); err != nil {
		log.Fatal(err)
	}

	checkW1()
	a, err := getApp(*cfg, &brewCfg)
	if err != nil {
		panic(err)
	}
	a.Start()
}

//I am too lazy to load the BB-W1 device tree overlay the right way, which would
//be to add it to uEnv.txt (which is a config file for u-boot).
func checkW1() {
	if !utils.FileExists("/sys/bus/w1/devices/28-0000047ade8f") {
		ioutil.WriteFile("/sys/devices/bone_capemgr.9/slots", []byte("BB-W1:00A0"), 0666)
	}
}

func getApp(cfg interface{}, brewCfg *brewery.BrewConfig) (*gogadgets.App, error) {
	hlt, tun, boiler, carboy := brewery.NewBrewery(brewCfg)
	a := gogadgets.NewApp(cfg, hlt, tun, boiler, carboy)
	return a, nil
}
