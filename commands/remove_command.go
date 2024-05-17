package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func HandleRemoveFileOrDirectoryCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	_, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
	if err := os.RemoveAll(filename); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)
}
