package recipes_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/cswank/brewery/recipes"
)

func TestExample(t *testing.T) {
	b, _ := ioutil.ReadFile("example.json")
	recipe := &recipes.Recipe{}
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

	recipe.WaterRatio = 1.25
	m := recipe.Method(75.0)
	if len(m) != 37 {
		t.Error(len(m))
	}
}
