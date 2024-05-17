package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func CreateDirectoryCommand(connection models.MessageHandler, message *models.Message, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	// get the requested directory path to create
	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(append([]byte{1}, []byte("bad path")...))
		return
	}

	// prepend the user root directory path
	filename = path.Join(clientFileDirectory, filename)
	if err := os.Mkdir(filename, os.ModePerm); err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}

	_ = connection.Write(append([]byte{0}, []byte("")...))
}
