package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) ValidateLogin(email, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	matchedUser, exists := dbStruct.Users[email]
	if !exists {
		return User{}, errors.New("User doesn't exist")
	}

	bytePassword := []byte(password)
	err = bcrypt.CompareHashAndPassword(matchedUser.Password, bytePassword)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:    matchedUser.ID,
		Email: matchedUser.Email,
	}, nil
}
