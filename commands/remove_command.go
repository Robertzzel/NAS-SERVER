package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func RemoveCommand(connection models.MessageHandler, message *models.MessageForServer, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	// get the requested file
	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	// prepend the user root directory
	filename = path.Join(clientFileDirectory, filename)
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
