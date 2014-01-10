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
	m := recipe.GetMethod()
	for _, s := range m {
		fmt.Println(s)
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
	m := recipe.getMash(25.0)
	if m.StrikeTemperature != 186.75 {
		t.Error(m.StrikeTemperature)
	}

	if recipe.getGrainWeight() != 16.999979782739803 {
		t.Error(recipe.getGrainWeight())
	}
}
