package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/johnmayou/uber/pkg/chassis"
)

func handler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "trip:"+r.URL.Path)
	_, _ = fmt.Fprintln(w, "add(1, 2):"+strconv.Itoa(chassis.Add(1, 2)))
}

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
