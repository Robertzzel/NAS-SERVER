package configurations

import (
	"os"
)

func GetServerHost() string {
	return os.Getenv("SERVER_HOST")
}

func GetServerPort() string {
	return os.Getenv("SERVER_PORT")
}

func GetFilesHost() string {
	return os.Getenv("FILES_SERVICE_HOST")
}

func GetFilesPort() string {
	return os.Getenv("FILES_SERVICE_PORT")
}

func GetDatabaseHost() string {
	return os.Getenv("DATABASE_HOST")
}

func GetDatabasePort() string {
	return os.Getenv("DATABASE_PORT")
}

func GetBaseFilesPath() string {
	return os.Getenv("STORAGE")
}
