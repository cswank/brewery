package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cswank/brewery/recipes"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	pth   = kingpin.Arg("input", "path to the recipe json file").Required().String()
	temp  = kingpin.Flag("temperature", "the temperature of the grains (F)").Short('t').Float()
	ratio = kingpin.Flag("ratio", "the grain/water ratio").Short('r').Default("1.25").Float()
)

func main() {
	kingpin.Parse()

	f, err := os.Open(*pth)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	r, err := recipes.New(f, recipes.WaterRatio(*ratio))
	if err != nil {
		log.Fatal(err)
	}

	m := r.Method(*temp)
	for _, row := range m {
		fmt.Println(row)
	}
}
