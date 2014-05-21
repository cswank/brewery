package controllers

import (
	"bitbucket.org/cswank/gadgetsweb/models"
	"bitbucket.org/cswank/brewery/recipes"
	"encoding/json"
	"net/http"
)


func GetPing(w http.ResponseWriter, r *http.Request, u *models.User, vars map[string]string) error {
	w.Write([]byte("ping"))
	return nil
}

func GetRecipe(w http.ResponseWriter, r *http.Request, u *models.User, vars map[string]string) error {
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


