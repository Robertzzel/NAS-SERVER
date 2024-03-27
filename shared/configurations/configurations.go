package configurations

import (
	"errors"
	"os"
)

var (
	host          string
	port          string
	databasePath  string
	baseFilesPath string
)

func UpdateConfigurations() error {
	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	if port == "" {
		return errors.New("no port given")
	}
	databasePath = os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		return errors.New("no database path given")
	}
	baseFilesPath = os.Getenv("STORAGE")
	if baseFilesPath == "" {
		return errors.New("no storage given")
	}

	return nil
}

func GetHost() string {
	return host
}

func GetPort() string {
	return port
}

func GetDatabasePath() string {
	return databasePath
}

func GetBaseFilesPath() string {
	return baseFilesPath
}
