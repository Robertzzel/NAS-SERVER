package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func RenameCommand(connection models.MessageHandler, message *models.MessageForServer, clientFileDirectory string) {
	if len(message.Args) != 2 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	// get file paths for the file and the new name/path of the file
	filename := message.Args[0]
	newFilename := message.Args[1]
	if !IsPathSafe(filename) && !IsPathSafe(newFilename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	// prepend both paths with user root directory
	filename = path.Join(clientFileDirectory, filename)
	newFilename = path.Join(clientFileDirectory, newFilename)

	if err := os.Rename(filename, newFilename); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)
}
