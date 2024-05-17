package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"strconv"
)

func InfoCommand(connection models.MessageHandler, message *models.MessageForServer, clientUsername string) {
	if len(message.Args) != 0 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	remainingMemory, err := services.GetUserRemainingMemory(clientUsername)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte(strconv.FormatInt(remainingMemory, 10))).Data)
}
