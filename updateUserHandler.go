package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding json")
		return
	}

	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.jwtSecret), nil
		},
	)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	expirationTime, err := jwtToken.Claims.GetExpirationTime()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token doesn't have an expiration")
		return
	}

	if expirationTime.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "Token is expired")
		return
	}

	userId, err := jwtToken.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Token doesn't contain a subject (ID)")
		return
	}

	resUser, err := cfg.database.UpdateUser(userId, params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	respondWithJSON(w, http.StatusOK, resUser)
}
