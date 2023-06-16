package apiconfig

import (
	"github.com/AxterDoesCode/webserver/internal/database"
)

type ApiConfig struct {
	FileserverHits int
	Database       database.DB
	JwtSecret      string
	PolkaKey       string
}
