package database

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
