package database

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/bcrypt"
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
		return dbstructure, nil
	}
	return dbstructure, nil
}

func (db *DB) AddUser(password, email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	_, exists, err := db.checkUserExists(email)
	if err != nil {
		return User{}, errors.New("Error checking user exists")
	}

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

	dbStruct.Users[dbNextIndex] = fullUserDetails
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}
	return returnUser, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	dbNextIndex := len(dbStruct.Chirps) + 1
	returnChirp := Chirp{
		ID:   dbNextIndex,
		Body: body,
	}

	dbStruct.Chirps[dbNextIndex] = returnChirp
	err = db.writeDB(dbStruct)
	if err != nil {
		return Chirp{}, err
	}
	return returnChirp, nil
}

func (db *DB) checkUserExists(email string) (User, bool, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, false, errors.New("failed to load DB")
	}
	for _, val := range dbStruct.Users {
		if val.Email == email {
			return val, true, nil
		}
	}
	return User{}, false, nil
}

func (db *DB) GetChirpByID(id int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	elem, ok := dbStruct.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("The chirp ID doesn't correspond to any Chirp")
	}
	return elem, nil
}

func (db *DB) GetChirpsArr() ([]Chirp, error) {
	dat, err := db.loadDB()
	chirpSlice := make([]Chirp, 0)
	if err != nil {
		return chirpSlice, err
	}

	for _, val := range dat.Chirps {
		chirpSlice = append(chirpSlice, val)
	}
	return chirpSlice, nil
}

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
	dbstructure.Users = make(map[int]User)

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