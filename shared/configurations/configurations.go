package configurations

import (
	"errors"
	"os"
)

var (
	serverHost    string
	serverPort    string
	databaseHost  string
	databasePort  string
	filesHost     string
	filesPort     string
	databasePath  string
	baseFilesPath string
)

func UpdateConfigurations() error {
	serverHost = os.Getenv("SERVER_HOST")
	serverPort = os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return errors.New("no port given")
	}
	databaseHost = os.Getenv("FILES_SERVICE_HOST")
	databasePort = os.Getenv("FILES_SERVICE_PORT")
	if databasePort == "" {
		return errors.New("no port given")
	}
	filesHost = os.Getenv("DATABASE_HOST")
	filesPort = os.Getenv("DATABASE_PORT")
	if filesPort == "" {
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

func GetServerHost() string {
	return serverHost
}

func GetServerPort() string {
	return serverPort
}

func GetFilesHost() string {
	return filesHost
}

func GetFilesPort() string {
	return filesPort
}

func GetDatabaseHost() string {
	return databaseHost
}

func GetDatabasePort() string {
	return databasePort
}

func GetDatabasePath() string {
	return databasePath
}

func GetBaseFilesPath() string {
	return baseFilesPath
}
