package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func HandleRenameFileOrDirectoryCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 2 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	filename := message.Args[0]
	newFilename := message.Args[1]
	if !IsPathSafe(filename) && !IsPathSafe(newFilename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	newFilename = path.Join(user.UserRootDirectory, newFilename)

	if err := os.Rename(filename, newFilename); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)
}
