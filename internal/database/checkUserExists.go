package database

import "errors"

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
