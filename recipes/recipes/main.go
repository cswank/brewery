package main

import (
	"fmt"
	"log"

	"bitbucket.org/cswank/brewery/recipes"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	name = kingpin.Flag("name", "the name of the brewtoad recipe").Short('n').String()
	temp = kingpin.Flag("temperature", "the temperature of the grains (F)").Short('t').Float()
)

func main() {
	kingpin.Parse()
	r, err := recipes.NewRecipe(*name)
	if err != nil {
		log.Fatal(err)
	}
	m := r.GetMethod(*temp)
	fmt.Println("[")
	for _, step := range m {
		fmt.Printf("  %s,\n", step)
	}
	fmt.Println("]")
}
