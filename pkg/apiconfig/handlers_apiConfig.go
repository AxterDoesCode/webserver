package apiconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/AxterDoesCode/webserver/pkg/httpHandler"
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
		return
	}

	chirp, err := cfg.Database.GetChirpByID(id)
	if err != nil {
		httphandler.RespondWithError(w, 404, "Chirp Doesn't exist")
		return
	}

	httphandler.RespondWithJSON(w, http.StatusOK, chirp)
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

	const maxLength = 140
	if len(params.Body) > maxLength {
		httphandler.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	tempChirp, err := cfg.Database.CreateChirp(cleanChirp(params.Body))
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

	expirationTime, err := jwtToken.Claims.GetExpirationTime()
	if err != nil {
		httphandler.RespondWithError(w, http.StatusUnauthorized, "Token doesn't have an expiration")
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

	jwtClaim := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: getExpirationTime(params.Expiration),
		Subject:   strconv.Itoa(user.ID),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaim)
	signedJwtToken, err := jwtToken.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		httphandler.RespondWithError(w, http.StatusInternalServerError, errStr)
		return
	}

	res := responseUser{
		ID:    user.ID,
		Email: user.Email,
		Token: signedJwtToken,
	}

	httphandler.RespondWithJSON(w, 200, res)
}

func getExpirationTime(s int) *jwt.NumericDate {
	secondsInHour := 60 * 60
	if s == 0 || s > secondsInHour {
		return jwt.NewNumericDate(time.Now().Add(24 * time.Hour).UTC())
	}
	return jwt.NewNumericDate(time.Now().Add(time.Duration(s) * time.Second).UTC())
}
