package models

import (
	"errors"
	"strings"
)

type RequestMessage struct {
	Command byte
	Args    []string
}

func NewRequestMessage(command byte, args []string) RequestMessage {
	return RequestMessage{
		Command: command,
		Args:    args,
	}
}

func NewRequestMessageFromBytes(command []byte) (RequestMessage, error) {
	rm := RequestMessage{
		Command: command[0],
	}

	if len(command) == 0 {
		return RequestMessage{}, errors.New("invalid message")
	}

	if len(command) > 1 {
		rm.Args = strings.Split(string(command[1:]), "\n")
	}

	return rm, nil
}

func (rm *RequestMessage) GetBytesData() []byte {
	msg := []byte{rm.Command}
	for _, arg := range rm.Args {
		msg = append(msg, []byte(arg)...)
	}
	return msg
}

type ResponseMessage struct {
	Status byte
	Body   []byte
}

func NewResponseMessage(status byte, body []byte) ResponseMessage {
	return ResponseMessage{
		Status: status,
		Body:   body,
	}
}

func NewResponseMessageFromBytes(message []byte) ResponseMessage {
	return ResponseMessage{
		Status: message[0],
		Body:   message[1:],
	}
}

func (rm *ResponseMessage) GetBytesData() []byte {
	msg := []byte{rm.Status}
	msg = append(msg, rm.Body...)
	return msg
}
