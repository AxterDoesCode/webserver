package database

import (
	"errors"
)

func (db *DB) GetChirpByID(id int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	elem, ok := dbStruct.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("The chirp ID doesn't correspond to any Chirp")
	}
	return elem, nil
}
