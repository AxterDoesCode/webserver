package database

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
