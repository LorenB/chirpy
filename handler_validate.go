package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	type returnCleanedVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode((&params))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	wordsToClean := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	cleanedBody := params.Body
	cleansed := "****"
	for _, word := range wordsToClean {
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(word))
		cleanedBody = re.ReplaceAllString(cleanedBody, cleansed)
	}

	respondWithJSON(w, http.StatusOK, returnCleanedVals{
		CleanedBody: cleanedBody,
	})
}
