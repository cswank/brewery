package main

import (
	"fmt"
	"flag"
	"net/http"
	"bitbucket.com/cswank/gadgetsweb/app"
)

var (
	static = flag.String("s", "", "Path to gogadget's static files")
)


func main() {
	flag.Parse()
	r := app.GetRouter(*static)
	http.Handle("/", r)
	fmt.Println("listening on 0.0.0.0:8080")
	http.ListenAndServe(":8080", nil)
}
