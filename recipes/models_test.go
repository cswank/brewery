package recipes

import (
	"testing"
	"io/ioutil"
	"encoding/xml"
)


func TestExample(t *testing.T) {
	b, _ := ioutil.ReadFile("example.xml")
	recipes := &Recipes{}
	err := xml.Unmarshal(b, recipes)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(recipes)
	if len(recipes.Recipes) != 1 {
		t.Fatal(recipes.Recipes)
	}
	recipe := recipes.Recipes[0]
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
	if f.Amount != 3.628734644520015 {
		t.Error(f)
	}
}
