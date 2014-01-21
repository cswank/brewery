package recipes

import (
	"testing"
	"io/ioutil"
	"encoding/json"
	"fmt"
)


func TestExample(t *testing.T) {
	b, _ := ioutil.ReadFile("example.json")
	recipe := &Recipe{}
	err := json.Unmarshal(b, recipe)
	if err != nil {
		t.Fatal(err)
	}
	if len(recipe.Fermentables) != 4 {
		t.Error(recipe.Fermentables)
	}
	if len(recipe.Hops) != 5 {
		t.Error(recipe.Hops)
	}
	if len(recipe.Yeasts) != 1 {
		t.Error(recipe.Yeasts)
	}

	f := recipe.Fermentables[0]
	if f.Name != "Caramel Malt 40L" {
		t.Error(f)
	}
	if f.Amount != 7.9999904859952 {
		t.Error(f)
	}
	
	if len(recipe.MashSteps) != 2 {
		t.Error(recipe.Yeasts)
	}
	
	step := recipe.MashSteps[0]
	if step.Temperature != 154.4 {
		t.Error(step)
	}
}

func TestMethod(t *testing.T) {
	b, _ := ioutil.ReadFile("example.json")
	recipe := &Recipe{}
	err := json.Unmarshal(b, recipe)
	if err != nil {
		t.Fatal(err)
	}
	recipe.WaterRatio = 1.25
	m := recipe.GetMethod(75.0)
	if len(m) != 37 {
		t.Error(len(m))
	}
	
}

func TestMash(t *testing.T) {
	b, _ := ioutil.ReadFile("example.json")
	recipe := &Recipe{}
	err := json.Unmarshal(b, recipe)
	if err != nil {
		t.Fatal(err)
	}
	recipe.WaterRatio = 1.25
	m := recipe.getMash(75.0)
	if m.StrikeTemperature != 174.25 {
		t.Error(m.StrikeTemperature)
	}

	if recipe.getGrainWeight() != 16.999979782739803 {
		t.Error(recipe.getGrainWeight())
	}
}


func TestNewRecipe(t *testing.T) {
	r, err := NewRecipe("vladimirs-own-stout")
	if err != nil {
		t.Fatal(err)
	}
	m := r.GetMethod(75.0)
	if m[1] != "heat hlt to 174.250000 F" {
		fmt.Println(m)
		t.Error(m[1])
	}
	if m[24] != "fill boiler to 8.437494 gallons" {
		t.Error(m[24])
	}
	for _, s := range r.GetMethod(75.0) {
		fmt.Println(s)
	}
}
