package database

import (
	"encoding/json"
	"errors"
	"os"
)

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	var dbstructure DBStructure
	file, err := os.ReadFile(db.path)
	if err != nil {
		return dbstructure, err
	}

	err = json.Unmarshal(file, &dbstructure)
	if err != nil {
		return dbstructure, errors.New("Error unmarshaling JSON")
	}
	return dbstructure, nil
}
