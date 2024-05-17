package configurations

import (
	"errors"
	"os"
)

var (
	Host          string
	Port          string
	DatabasePath  string
	BaseFilesBath string
)

func UpdateConfigurations() error {
	Host = os.Getenv("HOST")
	Port = os.Getenv("PORT")
	if Port == "" {
		return errors.New("no port given")
	}
	DatabasePath = os.Getenv("DATABASE_PATH")
	if DatabasePath == "" {
		return errors.New("no database path given")
	}
	BaseFilesBath = os.Getenv("STORAGE")
	if BaseFilesBath == "" {
		return errors.New("no storage given")
	}
	return nil
}
