package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.database.AddUser(params.Password, params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
}
