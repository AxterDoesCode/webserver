package apiconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/AxterDoesCode/webserver/pkg/httphandler"
)

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
	`, cfg.FileserverHits)))
}

func (cfg *ApiConfig) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "chirpID")

	id, err := strconv.Atoi(param)
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"error converting id to int",
		)
		return
	}

	chirp, err := cfg.Database.GetChirpByID(id)
	if err != nil {
		httphandler.RespondWithError(w, 404, "Chirp Doesn't exist")
		return
	}

	httphandler.RespondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *ApiConfig) DeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "chirpID")

	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		},
	)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	tokenIssuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if tokenIssuer != "chirpy-access" {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is not a refresh token")
		return
	}

	tokenSubject, err := jwtToken.Claims.GetSubject()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if param != tokenSubject {
		httphandler.RespondWithError(
			w,
			http.StatusForbidden,
			"Cannot Delete: Chirp does not belong to user",
		)
		return
	}

	paramID, err := strconv.Atoi(param)
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"error converting id to int",
		)
		return
	}
	deletedChirp, err := cfg.Database.DeleteChirpByID(paramID)
	httphandler.RespondWithJSON(w, http.StatusOK, deletedChirp)
}

func (cfg *ApiConfig) AddUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't decode parameters",
		)
		return
	}

	user, err := cfg.Database.AddUser(params.Password, params.Email)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	httphandler.RespondWithJSON(w, http.StatusCreated, user)
}

func (cfg *ApiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	slice, err := cfg.Database.GetChirpsArr()
	if err != nil {
		return
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].ID < slice[j].ID
	})
	httphandler.RespondWithJSON(w, http.StatusOK, slice)
}

func (cfg *ApiConfig) PostChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		},
	)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	tokenIssuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if tokenIssuer != "chirpy-access" {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is not an access token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't decode parameters",
		)
		return
	}

	const maxLength = 140
	if len(params.Body) > maxLength {
		httphandler.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	claimSubject, err := jwtToken.Claims.GetSubject()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	id, err := strconv.Atoi(claimSubject)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	tempChirp, err := cfg.Database.CreateChirp(id, cleanChirp(params.Body))
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		log.Print(err)
		return
	}
	httphandler.RespondWithJSON(w, 201, tempChirp)
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

func (cfg *ApiConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, "Error decoding json")
		return
	}

	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		},
	)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	tokenIssuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if tokenIssuer == "chirpy-refresh" {
		httphandler.RespondWithError(
			w,
			http.StatusUnauthorized,
			"Provided token is a refresh token",
		)
		return
	}

	expirationTime, err := jwtToken.Claims.GetExpirationTime()
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusUnauthorized,
			"Token error getting token expirationTime",
		)
		return
	}

	if expirationTime.Before(time.Now().UTC()) {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is expired")
		return
	}

	userId, err := jwtToken.Claims.GetSubject()
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Token doesn't contain a subject (ID)",
		)
		return
	}

	resUser, err := cfg.Database.UpdateUser(userId, params.Email, params.Password)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	httphandler.RespondWithJSON(w, http.StatusOK, resUser)
}

func (cfg *ApiConfig) UserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"Password"`
	}

	type responseUser struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		ChirpyRed    bool   `json:"is_chirpy_red"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to decode request, expected json",
		)
		return
	}

	user, err := cfg.Database.ValidateLogin(params.Email, params.Password)
	if err != nil {
		httphandler.RespondWithError(w, 401, "Unauthorized")
		return
	}

	// Creating Access and Refresh token
	signedJwtAccessToken, err := generateJwtToken(user.ID, "chirpy-access", cfg.JwtSecret)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	signedJwtRefreshToken, err := generateJwtToken(user.ID, "chirpy-refresh", cfg.JwtSecret)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	res := responseUser{
		ID:           user.ID,
		Email:        user.Email,
		Token:        signedJwtAccessToken,
		RefreshToken: signedJwtRefreshToken,
		ChirpyRed:    user.ChirpyRed,
	}

	httphandler.RespondWithJSON(w, 200, res)
}

func generateJwtToken(id int, issuer, secret string) (string, error) {
	var claimIssuer string
	var expirationTime *jwt.NumericDate
	switch issuer {
	case "chirpy-access":
		claimIssuer = "chirpy-access"
		expirationTime = jwt.NewNumericDate(time.Now().Add(1 * time.Hour).UTC())
	case "chirpy-refresh":
		claimIssuer = "chirpy-refresh"
		expirationTime = jwt.NewNumericDate(time.Now().Add(60 * 24 * time.Hour).UTC())
	default:
		return "", errors.New("Issuer string isn't valid")
	}

	jwtClaim := jwt.RegisteredClaims{
		Issuer:    claimIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: expirationTime,
		Subject:   strconv.Itoa(id),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaim)
	signedJwtToken, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		return "", errors.New(errStr)
	}
	return signedJwtToken, nil
}

func (cfg *ApiConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	type returnToken struct {
		Token string `json:"token"`
	}
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		},
	)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	tokenIssuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if tokenIssuer != "chirpy-refresh" {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is not a refresh token")
		return
	}

	err = cfg.Database.CheckRefreshTokenRevoked(tokenString)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is revoked")
		return
	}

	idStr, err := jwtToken.Claims.GetSubject()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		httphandler.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Error converting Token subject ID to int",
		)
		return
	}

	returnAccessToken, err := generateJwtToken(id, "chirpy-access", cfg.JwtSecret)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	returnAccessTokenStruct := returnToken{
		Token: returnAccessToken,
	}
	httphandler.RespondWithJSON(w, http.StatusOK, returnAccessTokenStruct)
}

func (cfg *ApiConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		},
	)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	tokenIssuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if tokenIssuer != "chirpy-refresh" {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token is not a refresh token")
		return
	}

	err = cfg.Database.CheckRefreshTokenRevoked(tokenString)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusAlreadyReported, "Token is already revoked")
		return
	}

	returnToken, err := cfg.Database.RevokeToken(tokenString, time.Now())
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}
	httphandler.RespondWithJSON(w, http.StatusOK, returnToken)
}

func (cfg *ApiConfig) UserUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	requestParams := params{}

	err := decoder.Decode(&requestParams)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%s", err))
		return
	}

	if requestParams.Event != "user.upgraded" {
		httphandler.RespondWithJSON(w, http.StatusOK, params{})
		return
	}
	err = cfg.Database.UpgradeUser(requestParams.Data.UserID)
	if err != nil {
		httphandler.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("%s", err))
		return
	}
	httphandler.RespondWithJSON(w, http.StatusOK, params{})
}
