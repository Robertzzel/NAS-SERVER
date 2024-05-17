package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func RemoveCommand(connection models.MessageHandler, message *models.Message, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	// get the requested file
	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(append([]byte{1}, []byte("bad path")...))
		return
	}

	// prepend the user root directory
	filename = path.Join(clientFileDirectory, filename)
	_, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}
	if err := os.RemoveAll(filename); err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}

	_ = connection.Write(append([]byte{0}, []byte("")...))
}
