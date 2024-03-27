package UserService

import (
	"NAS-Server-Web/DatabaseService"
)

var instance *UserService = nil

type UserService struct {
	db DatabaseService.DatabaseService
}

func NewUserService() (UserService, error) {
	if instance == nil {
		db, err := DatabaseService.NewDatabaseService()
		if err != nil {
			return UserService{}, err
		}

		dm := UserService{db}
		instance = &dm
	}

	return UserService{}, nil
}

func (service *UserService) CheckUsernameAndPassword(username, password string) (bool, error) {
	return service.db.UsernameAndPasswordExists(username, password)
}

func (service *UserService) GetUserAllocatedMemory(username string) (int, error) {
	return service.GetUserAllocatedMemory(username)
}
