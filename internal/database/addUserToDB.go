package database

func (db *DB) AddUser(email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	dbNextIndex := len(dbStruct.Users) + 1
	returnUser := User{
		ID:    dbNextIndex,
		Email: email,
	}

	dbStruct.Users[email] = returnUser
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}
	return returnUser, nil
}
