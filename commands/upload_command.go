package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"os"
	"path"
	"strconv"
)

func UploadCommand(connection models.MessageHandler, message *models.Message, clientUsername string, userFileDirectory string) {
	if len(message.Args) != 2 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	// get file path and size
	filename := message.Args[0]
	size, err := strconv.Atoi(message.Args[1])
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("invalid size")...))
		return
	}

	// check if the user has enough memory
	remainingMemory, err := services.GetUserRemainingMemory(clientUsername)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}
	if remainingMemory < int64(size) {
		_ = connection.Write(append([]byte{1}, []byte("no memory for the upload")...))
		return
	}

	if !IsPathSafe(filename) {
		_ = connection.Write(append([]byte{1}, []byte("bad path")...))
		return
	}

	// prepend the filepath with the user root directory
	filename = path.Join(userFileDirectory, filename)
	file, err := os.Create(filename)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}
	defer file.Close()

	// send confirmation message so that the client knows it can send the file contents
	_ = connection.Write(append([]byte{0}, []byte("")...))

	if err = connection.ReadFile(file); err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}

	_ = connection.Write(append([]byte{0}, []byte("")...))
}
