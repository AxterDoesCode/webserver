package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email      string `json:"email"`
		Password   string `json:"Password"`
		Expiration int    `json:"expires_in_seconds"`
	}

	type responseUser struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
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

	jwtClaim := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: getExpirationTime(params.Expiration),
		Subject:   strconv.Itoa(user.ID),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaim)
	signedJwtToken, err := jwtToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		respondWithError(w, http.StatusInternalServerError, errStr)
		return
	}

	res := responseUser{
		ID:    user.ID,
		Email: user.Email,
		Token: signedJwtToken,
	}

	respondWithJSON(w, 200, res)
}

func getExpirationTime(s int) *jwt.NumericDate {
	secondsInHour := 60 * 60
	if s == 0 || s > secondsInHour {
		return jwt.NewNumericDate(time.Now().Add(24 * time.Hour).UTC())
	}
	return jwt.NewNumericDate(time.Now().Add(time.Duration(s) * time.Second).UTC())
}
