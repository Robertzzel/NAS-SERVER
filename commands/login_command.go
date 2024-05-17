package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"path/filepath"
)

func HandleLoginCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	config, err := services.NewConfigsService()
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if len(message.Args) != 2 {
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
	if exists {
		user.IsAuthenticated = true
		user.Name = username
		user.UserRootDirectory = filepath.Join(config.BaseFilesBath, username)
	} else {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid username or password")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)
}
