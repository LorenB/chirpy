package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode((&params))
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	type returnVals struct {
		Valid bool `json:"valid"`
	}
	respBody := returnVals{
		Valid: true,
	}

	type errorResp struct {
		Error string `json:"error"`
	}
	errorBody := errorResp{
		Error: "Chirp is too long",
	}
	errorData, err := json.Marshal(errorBody)
	if err != nil {
		log.Printf("Error marhalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(400)
		w.Write(errorData)
		return
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marhalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}
