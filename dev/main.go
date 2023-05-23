package main

import (
	"log"
	"net/http"

	"github.com/burstsms/hack-making-it-known"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", mik.Handler)

	log.Print("Listening...")
	err := http.ListenAndServe(":3000", mux)
	if err != nil {
		return
	}

}
