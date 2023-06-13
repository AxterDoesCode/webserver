package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) AddUser(password, email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	_, exists := dbStruct.Users[email]
	if exists {
		return User{}, errors.New("User already exists, please try a different email or login")
	}

	dbNextIndex := len(dbStruct.Users) + 1
	returnUser := User{
		ID:    dbNextIndex,
		Email: email,
	}

	fullUserDetails := returnUser

	bytePass := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(bytePass, bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	fullUserDetails.Password = hash

	dbStruct.Users[email] = fullUserDetails
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}
	return returnUser, nil
}
