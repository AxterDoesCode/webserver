package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "chirpID")

	id, err := strconv.Atoi(param)
	if err != nil {
		return
	}

	chirp, err := cfg.database.GetChirpByID(id)
	if err != nil {
		respondWithError(w, 404, "Chirp Doesn't exist")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
