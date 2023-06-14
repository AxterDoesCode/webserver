package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) ValidateLogin(email, password string) (User, error) {
	matchedUser, exists, err := db.checkUserExists(email)
	if !exists {
		return User{}, errors.New("User doesn't exist")
	}

	if err != nil {
		return User{}, errors.New("Error checking user exists")
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
