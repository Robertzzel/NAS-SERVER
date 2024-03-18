package models

import (
	"errors"
	"strings"
)

type MessageForServer struct {
	Command byte
	Args    []string
}

func NewMessage(command []byte) (MessageForServer, error) {
	message := MessageForServer{}
	message.Command = command[0]

	if len(command) == 0 {
		return MessageForServer{}, errors.New("invalid message")
	}

	if len(command) > 1 {
		message.Args = strings.Split(string(command[1:]), "\n")
	}

	return message, nil
}
