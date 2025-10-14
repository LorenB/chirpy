package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = mux

	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}
