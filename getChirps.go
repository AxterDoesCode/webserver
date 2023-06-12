package main

import (
	"net/http"
	"sort"
)

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	slice, err := cfg.database.GetChirpsArr()
	if err != nil {
		return
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].ID < slice[j].ID
	})
	respondWithJSON(w, http.StatusOK, slice)
}
