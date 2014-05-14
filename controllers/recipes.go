package controllers

import (
	"github.com/gorilla/mux"
	"bitbucket.org/cswank/gadgetsweb/models"
	"bitbucket.org/cswank/brewery/recipes"
	"encoding/json"
	"net/http"
)

func GetRecipe(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	recipeName := vars["name"]
	recipe, err := recipes.NewRecipe(recipeName)
	if err != nil {
		return err
	}
	steps := recipe.GetMethod(70.0)
	method := &models.Method{
		Gadget: "brewery",
		Steps: steps,
	}
	b, err := json.Marshal(method)
	if err != nil {
		return err
	}
	w.Write(b)
	return err
}


