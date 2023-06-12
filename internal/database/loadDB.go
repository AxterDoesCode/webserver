package database

import (
	"encoding/json"
	"os"
)

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	var dbstructure DBStructure
	dbstructure.Chirps = make(map[int]Chirp)

	file, err := os.ReadFile(db.path)
	if err != nil {
		return dbstructure, err
	}

	err = json.Unmarshal(file, &dbstructure)
	if err != nil {
		return dbstructure, nil
	}
	return dbstructure, nil
}
