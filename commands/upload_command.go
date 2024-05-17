package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func HandleUploadCommand(connection *models.MessageHandler, message *models.MessageForServer) {
	if len(message.Args) != 4 {
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
	size, err := strconv.Atoi(message.Args[3])
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid size")).Data)
		return
	}

	remainingMemory, err := services.GetUserRemainingMemory(username)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if remainingMemory < int64(size) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("no memory for the upload")).Data)
		return
	}

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
	file, err := os.Create(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
	defer file.Close()

	_ = connection.Write(models.NewMessageForClient(0, []byte("go on")).Data)

	if err = connection.ReadFile(file); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)
}
