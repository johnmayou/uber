package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/johnmayou/uber/pkg/chassis"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "auth:"+r.URL.Path)
	fmt.Fprintln(w, "add(1, 1):"+strconv.Itoa(chassis.Add(1, 1)))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
