package commands

import (
	"NAS-Server-Web/services"
	"errors"
	"strings"
)

const (
	UploadFile              = 0
	DownloadFileOrDirectory = 1
	CreateDirectory         = 2
	RemoveFileOrDirectory   = 3
	RenameFileOrDirectory   = 4
	Login                   = 5
	ListFilesAndDirectories = 6
	Info                    = 7
)

func IsPathSafe(path string) bool {
	return !strings.Contains(path, "../")
}

func checkUsernameAndPassword(name, password string) (bool, error) {
	db, err := services.NewDatabaseService()
	if err != nil {
		return false, errors.New("internal error")
	}

	return db.UsernameAndPasswordExists(name, password)
}
