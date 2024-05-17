package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func RenameCommand(connection models.MessageHandler, message *models.Message, clientFileDirectory string) {
	if len(message.Args) != 2 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	// get file paths for the file and the new name/path of the file
	filename := message.Args[0]
	newFilename := message.Args[1]
	if !IsPathSafe(filename) && !IsPathSafe(newFilename) {
		_ = connection.Write(append([]byte{1}, []byte("bad path")...))
		return
	}

	// prepend both paths with user root directory
	filename = path.Join(clientFileDirectory, filename)
	newFilename = path.Join(clientFileDirectory, newFilename)

	if err := os.Rename(filename, newFilename); err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}

	_ = connection.Write(append([]byte{0}, []byte("")...))
}
