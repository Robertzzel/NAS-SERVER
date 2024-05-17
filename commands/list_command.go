package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"path"
	"strconv"
)

func ListCommand(connection models.MessageHandler, message *models.Message, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	// get the requested directory path
	directoryPath := message.Args[0]
	if !IsPathSafe(directoryPath) {
		_ = connection.Write(append([]byte{1}, []byte("bad path")...))
		return
	}

	// prepend the user root directory path
	directoryPath = path.Join(clientFileDirectory, directoryPath)
	directory, err := services.GetFilesFromDirectory(directoryPath)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
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

	_ = connection.Write(append([]byte{0}, []byte(resultMessage)...))
}
