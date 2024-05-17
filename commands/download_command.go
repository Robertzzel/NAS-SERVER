package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"os"
	"path"
	"path/filepath"
)

func HandleDownloadFileOrDirectory(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if len(message.Args) != 3 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := checkUsernameAndPassword(username, password)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte(err.Error())).Data)
		return
	}
	if !exists {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	filename := message.Args[2]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	config, err := services.NewConfigsService()
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	userRootDirectory := filepath.Join(config.BaseFilesBath, username)
	filename = path.Join(userRootDirectory, filename)
	stat, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if stat.IsDir() {
		_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)

		err := connection.SendDirectoryAsZip(filename, user.UserRootDirectory)
		if err != nil {
			_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
			return
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
			return
		}
		defer file.Close()

		_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)

		if err = connection.SendFile(file); err != nil {
			_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
			return
		}
	}
}
