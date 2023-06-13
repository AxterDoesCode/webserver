package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"Password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to decode request, expected json",
		)
		return
	}
	user, err := cfg.database.ValidateLogin(params.Email, params.Password)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	respondWithJSON(w, 200, user)
}
