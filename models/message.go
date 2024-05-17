package models

import (
	"errors"
	"strings"
)

type Message struct {
	Command byte
	Args    []string
}

func NewMessage(command []byte) (Message, error) {
	message := Message{}
	message.Command = command[0]

	if len(command) == 0 {
		return Message{}, errors.New("invalid message")
	}

	if len(command) > 1 {
		message.Args = strings.Split(string(command[1:]), "\n")
	}

	return message, nil
}
