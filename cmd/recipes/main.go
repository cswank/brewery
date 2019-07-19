package main

import (
	"fmt"
	"log"

	"github.com/cswank/brewery/recipes"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	pth  = kingpin.Arg("input", "path to the recipe json file").Required().String()
	temp = kingpin.Flag("temperature", "the temperature of the grains (F)").Short('t').Float()
)

func main() {
	kingpin.Parse()
	r, err := recipes.NewRecipe(*pth)
	if err != nil {
		log.Fatal(err)
	}
	m := r.GetMethod(*temp)
	for _, row := range m {
		fmt.Println(row)
	}
}
