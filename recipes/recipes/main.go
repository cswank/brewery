package main

import (
	"fmt"
	"log"
	"encoding/json"

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
	b, _ := json.MarshalIndent(m, "", "  ")
	fmt.Println(string(b))
}
