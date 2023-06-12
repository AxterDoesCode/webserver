package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (cfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxLength = 140
	if len(params.Body) > maxLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	tempChirp, err := cfg.database.CreateChirp(cleanChirp(params.Body))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		log.Print(err)
		return
	}
	respondWithJSON(w, 201, tempChirp)
}

func cleanChirp(text string) string {
	chirpSlice := strings.Split(text, " ")

	for i, word := range chirpSlice {
		loweredWord := strings.ToLower(word)
		if loweredWord == "kerfuffle" || loweredWord == "sharbert" || loweredWord == "fornax" {
			chirpSlice[i] = "****"
		}
	}
	return strings.Join(chirpSlice, " ")
}
