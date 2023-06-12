package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	file, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatal(err)
	}

	var dbstructure DBStructure
	err = json.Unmarshal(file, &dbstructure)
	if err != nil {
		return dbstructure, errors.New("Error unmarshaling JSON")
	}
	return dbstructure, nil
}
