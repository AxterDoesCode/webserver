package database

import (
	"log"
	"os"
)

func NewDB(path string) (*DB, error) {
	_, err := os.Create(path + "/database.json")
	if err != nil {
		log.Fatal(err)
	}
	returnDB := DB{
		path: path + "/database.json",
	}
	return &returnDB, nil
}
