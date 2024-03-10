package services

import (
	"errors"
	"os"
)

const (
	configFile = "./configs.json"
)

var (
	serviceInstance *ConfigsService = nil
)

type ConfigsService struct {
	Host          string
	Port          string
	DatabasePath  string
	BaseFilesBath string
}

func NewConfigsService() (*ConfigsService, error) {
	if serviceInstance == nil {
		serviceInstance = &ConfigsService{}
		serviceInstance.Host = os.Getenv("HOST")
		serviceInstance.Port = os.Getenv("PORT")
		if serviceInstance.Port == "" {
			return nil, errors.New("no port given")
		}
		serviceInstance.DatabasePath = os.Getenv("DATABASE_PATH")
		if serviceInstance.DatabasePath == "" {
			return nil, errors.New("no database path given")
		}
		serviceInstance.BaseFilesBath = os.Getenv("STORAGE")
		if serviceInstance.BaseFilesBath == "" {
			return nil, errors.New("no storage given")
		}

	}

	return serviceInstance, nil
}
