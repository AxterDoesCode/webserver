package database

import (
	"sync"
	"time"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"Password,omitempty"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type RevokedToken struct {
	ID         string    `json:"id"`
	RevokeTime time.Time `json:"revoke_time"`
}

type DBStructure struct {
	Chirps        map[int]Chirp           `json:"chirps"`
	Users         map[int]User            `json:"users"`
	RevokedTokens map[string]RevokedToken `json:"revoked_tokens"`
}
