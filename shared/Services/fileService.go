package Services

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"crypto/tls"
	"errors"
	"strconv"
)

func GetUserUsedMemory(username string) (int64, error) {
	config, err := shared.GetTLSConfigs()
	if err != nil {
		return 0, err
	}
	address := configurations.GetFilesHost() + ":" + configurations.GetDatabasePort()
	conn, err := tls.Dial("tcp", address, config)
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
	config, err := shared.GetTLSConfigs()
	if err != nil {
		return "", err
	}
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	conn, err := tls.Dial("tcp", address, config)
	if err != nil {
		return "", err
	}
	mh := shared.NewMessageHandler(conn)

	request := models.NewRequestMessage(3, []string{path})
	_ = mh.Write(request.GetBytesData())

	rawMsg, err := mh.Read()
	if err != nil {
		return "", err
	}

	response := models.NewResponseMessageFromBytes(rawMsg)
	return string(response.Body), nil
}

func Download(path string) error {
	return nil
}

func Upload(path string) error {
	return nil
}

func CreateDirectory(path string) error {
	config, err := shared.GetTLSConfigs()
	if err != nil {
		return err
	}
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	conn, err := tls.Dial("tcp", address, config)
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
	config, err := shared.GetTLSConfigs()
	if err != nil {
		return err
	}
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	conn, err := tls.Dial("tcp", address, config)
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
	config, err := shared.GetTLSConfigs()
	if err != nil {
		return err
	}
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	conn, err := tls.Dial("tcp", address, config)
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
