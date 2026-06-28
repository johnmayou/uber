package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "auth:"+r.URL.Path)
	})
	http.ListenAndServe(":8080", nil)
}
