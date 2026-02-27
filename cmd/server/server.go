package main

import (
	"log"
	"net/http"
)

func main() {
	const addr = ":8080"

	fs := http.FileServer(http.Dir("."))

	log.Printf("Serving at port %s", addr)

	err := http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		fs.ServeHTTP(w, r)
	}))

	if err != nil {
		log.Fatal(err)
	}
}
