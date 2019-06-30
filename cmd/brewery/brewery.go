package main

import (
	"flag"
	"log"

	"github.com/cswank/brewery"
	"github.com/cswank/gogadgets"
	"github.com/kelseyhightower/envconfig"
)

var (
	cfg = flag.String("c", "", "Path to the gogadgets config json file")
)

func main() {
	flag.Parse()
	var brewCfg brewery.Config
	if err := envconfig.Process("brewery", &brewCfg); err != nil {
		log.Fatal(err)
	}

	a, err := getApp(*cfg, &brewCfg)
	if err != nil {
		panic(err)
	}
	a.Start()
}

func getApp(cfg interface{}, brewCfg *brewery.Config) (*gogadgets.App, error) {
	hlt, tun, boiler, carboy, err := brewery.New(brewCfg)
	if err != nil {
		return nil, err
	}

	return gogadgets.New(cfg, hlt, tun, boiler, carboy), nil
}
