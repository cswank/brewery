package main

import (
	"log"
	"net/http"
	"bitbucket.org/cswank/brewery/controllers"
	"bitbucket.org/cswank/gadgetsweb/auth"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/recipes/_ping", GetPing).Methods("GET")
	r.HandleFunc("/recipes/{name}", GetRecipe).Methods("GET")
	http.Handle("/", r)
	log.Println("listening on 0.0.0.0:8081")
	http.ListenAndServe(":8081", nil)
}

func GetRecipe(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.GetRecipe)
}

func GetPing(w http.ResponseWriter, r *http.Request, "read") {
	auth.CheckAuth(w, r, controllers.GetPing)
}
