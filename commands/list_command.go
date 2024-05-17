package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"path"
	"strconv"
)

func ListCommand(connection models.MessageHandler, message *models.MessageForServer, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	// get the requested directory path
	directoryPath := message.Args[0]
	if !IsPathSafe(directoryPath) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	// prepend the user root directory path
	directoryPath = path.Join(clientFileDirectory, directoryPath)
	directory, err := services.GetFilesFromDirectory(directoryPath)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	//format the message
	resultMessage := ""
	for file := range directory {
		resultMessage += directory[file].Name + "\n" + strconv.FormatInt(directory[file].Size, 10) + "\n" + strconv.FormatBool(directory[file].IsDir) + "\n" + directory[file].Type + "\n" + strconv.FormatInt(directory[file].Created, 10) + "\x1c"
	}
	if len(resultMessage) > 0 {
		resultMessage = resultMessage[:len(resultMessage)-1]
	}

	if err := connection.Write(models.NewMessageForClient(0, []byte(resultMessage)).Data); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
}
