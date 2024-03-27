package UserService

import "NAS-Server-Web/DatabaseService"

type UserService struct {
	db DatabaseService.DatabaseService
}

func (service *UserService) CheckUsernameAndPassword(username, password string) (bool, error) {
	return service.db.UsernameAndPasswordExists(username, password)
}

func (service *UserService) GetUserAllocatedMemory(username string) (int, error) {
	return service.GetUserAllocatedMemory(username)
}
