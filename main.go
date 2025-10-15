package main

import (
	"log"
	"net/http"
)

type MyRootHanlder struct{}

// Implements the http.Handler interface directly in the struct.
// This method's signature exactly matches the interface requirement.
func (h *MyRootHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// write contents of index.html
	// http.FileServer(http.Dir("."))
	http.ServeFile(w, r, "index.html")
	// fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}

func main() {
	mux := http.NewServeMux()
	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = mux

	// 1. Create instance of the struct.
	// The variable 'handler' is now a value of type *MyRootHandler,
	// and it statifies the http.Handler interface.
	handler := &MyRootHanlder{}

	mux.Handle("/", handler)
	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}
