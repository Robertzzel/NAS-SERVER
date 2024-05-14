package Services

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"errors"
	"io"
	"net"
	"strconv"
)

func GetUserUsedMemory(username string) (int64, error) {
	conn, err := GetFileServerConnection()
	if err != nil {
		return 0, err
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(3, []string{username})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return 0, err
	}

	response := models.NewResponseMessageFromBytes(rawMsg)
	return strconv.ParseInt(string(response.Body), 10, 64)
}

func GetFilesFromDirectory(path string) (string, error) {
	conn, err := GetFileServerConnection()
	if err != nil {
		return "cannot get file server connection", err
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(2, []string{path})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return "cannot read the server response", err
	}

	response := models.NewResponseMessageFromBytes(rawMsg)
	return string(response.Body), nil
}

func Download(path string, clientC *shared.MessageHandler) {
	conn, err := GetFileServerConnection()
	if err != nil {
		return
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(0, []string{path})
	_ = mh.Write(request.GetBytesData())

	io.Copy(clientC.Conn, conn)
	conn.Close()
}

func Upload(path string, clientMh *shared.MessageHandler) {
	conn, err := GetFileServerConnection()
	if err != nil {
		return
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(1, []string{path})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return
	}

	_ = clientMh.Write(rawMsg)
	//TODO PROBLEME DOAR LA UPLOAD

	io.Copy(clientMh.Conn, conn)
}

func CreateDirectory(path string) error {
	conn, err := GetFileServerConnection()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(4, []string{path})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return err
	}

	response := models.NewResponseMessageFromBytes(rawMsg)
	if response.Status == 1 {
		return errors.New("cannot create the directory")
	}
	return nil
}

func RenameFileOrDirectory(fullPath, newFullPath string) error {
	conn, err := GetFileServerConnection()
	if err != nil {
		return err
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(5, []string{fullPath, newFullPath})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return err
	}

	response := models.NewResponseMessageFromBytes(rawMsg)
	if response.Status == 1 {
		return errors.New("cannot rename the file or directory")
	}
	return nil
}

func DeleteFileOrDirectory(fullPath string) error {
	conn, err := GetFileServerConnection()
	if err != nil {
		return err
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(6, []string{fullPath})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return err
	}

	response := models.NewResponseMessageFromBytes(rawMsg)
	if response.Status == 1 {
		return errors.New("cannot delete the directory")
	}
	return nil
}

func GetFileServerConnection() (net.Conn, error) {
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
