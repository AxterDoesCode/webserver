package database

import (
	"encoding/json"
	"os"
)

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	marshalData, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, marshalData, 0666)
	if err != nil {
		return err
	}
	return nil
}
