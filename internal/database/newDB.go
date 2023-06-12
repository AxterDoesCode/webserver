package database

import (
	"encoding/json"
	"os"
	"sync"
)

func NewDB(path string) (*DB, error) {
	_, err := os.Create(path + "/database.json")
	if err != nil {
		return nil, err
	}

	returnDB := DB{
		path: path + "/database.json",
		mux:  &sync.RWMutex{},
	}

	var dbstructure DBStructure
	dbstructure.Chirps = make(map[int]Chirp)

	dat, err := json.Marshal(dbstructure)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(returnDB.path, dat, 0666)
	if err != nil {
		return nil, err
	}
	return &returnDB, nil
}
