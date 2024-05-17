package models

type MessageForClient struct {
	Data []byte
}

func NewMessageForClient(result byte, message []byte) MessageForClient {
	return MessageForClient{append([]byte{result}, message...)}
}
