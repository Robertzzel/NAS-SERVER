package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"strconv"
)

func InfoCommand(connection models.MessageHandler, message *models.Message, clientUsername string) {
	if len(message.Args) != 0 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	remainingMemory, err := services.GetUserRemainingMemory(clientUsername)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}

	_ = connection.Write(append([]byte{0}, []byte(strconv.FormatInt(remainingMemory, 10))...))
}
