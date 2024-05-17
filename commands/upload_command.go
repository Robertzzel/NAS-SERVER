package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"os"
	"path"
	"strconv"
)

func UploadCommand(connection models.MessageHandler, message *models.MessageForServer, clientUsername string, userFileDirectory string) {
	if len(message.Args) != 2 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	// get file path and size
	filename := message.Args[0]
	size, err := strconv.Atoi(message.Args[1])
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid size")).Data)
		return
	}

	// check if the user has enough memory
	remainingMemory, err := services.GetUserRemainingMemory(clientUsername)
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

	// prepend the filepath with the user root directory
	filename = path.Join(userFileDirectory, filename)
	file, err := os.Create(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
	defer file.Close()

	// send confirmation message so that the client knows it can send the file contents
	_ = connection.Write(models.NewMessageForClient(0, []byte("go on")).Data)

	if err = connection.ReadFile(file); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)
}
