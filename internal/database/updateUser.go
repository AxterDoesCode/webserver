package database

import (
	"errors"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) UpdateUser(idStr, email, newPassword string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return User{}, err
	}

	elem, ok := dbStruct.Users[id]
	if !ok {
		return User{}, errors.New("ID cannot be found in databse")
	}

	bytePass := []byte(newPassword)
	hash, err := bcrypt.GenerateFromPassword(bytePass, bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	elem.Email = email
	elem.Password = hash
	dbStruct.Users[id] = elem
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	returnUser := User{
		ID:    elem.ID,
		Email: elem.Email,
	}

	return returnUser, nil
}
