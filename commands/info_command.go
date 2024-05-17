package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"strconv"
)

func HandleInfoCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 0 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	remainingMemory, err := services.GetUserRemainingMemory(user.Name)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte(strconv.FormatInt(remainingMemory, 10))).Data)
}
