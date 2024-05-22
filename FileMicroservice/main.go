package main

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DOWNLOAD    = 0
	UPLOAD      = 1
	LIST        = 2
	USED_MEMORY = 3
	CREATE      = 4
	RENAME      = 5
	DELETE      = 6
)

func main() {
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	log.Printf("Starting at " + address + "...")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}

		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()
	connection := shared.NewMessageHandler(c)

	rawMessage, err := connection.Read()
	if err != nil {
		log.Print("Closed connection with ", c.RemoteAddr())
		return
	}

	message, err := models.NewRequestMessageFromBytes(rawMessage)
	if err != nil {
		log.Print("Bad message structure from ", c.RemoteAddr())
		return
	}

	switch message.Command {
	case UPLOAD:
		if len(message.Args) != 1 {
			return
		}

		filePath := message.Args[0]
		file, err := os.Create(filePath)
		if err != nil {
			_ = SendResponseMessage(connection, 1, "internal error")
			return
		}

		response := models.NewResponseMessage(0, []byte("success"))
		if err := connection.Write(response.GetBytesData()); err != nil {
			return
		}
		_ = connection.ReadFile(file)
		_ = file.Close()
		log.Printf("Fisier scris, confirmare trimisa")
	case DOWNLOAD:
		if len(message.Args) != 1 {
			return
		}

		filePath := message.Args[0]
		file, err := os.Open(filePath)
		if err != nil {
			_ = SendResponseMessage(connection, 1, "file does not exist")
			return
		}

		response := models.NewResponseMessage(0, []byte(""))
		if err := connection.Write(response.GetBytesData()); err != nil {
			return
		}
		_ = connection.SendFile(file)
		_ = file.Close()
	case LIST:
		if len(message.Args) != 1 {
			return
		}

		directoryPath := message.Args[0]

		directory, err := GetFilesFromDirectory(directoryPath)
		if err != nil {
			return
		}

		resultMessage := ""
		for file := range directory {
			resultMessage += directory[file].Name + "\n" + strconv.FormatInt(directory[file].Size, 10) + "\n" + strconv.FormatBool(directory[file].IsDir) + "\n" + directory[file].Type + "\n" + strconv.FormatInt(directory[file].Created, 10) + "\x1c"
		}
		if len(resultMessage) > 0 {
			resultMessage = resultMessage[:len(resultMessage)-1]
		}

		response := models.NewResponseMessage(0, []byte(resultMessage))
		if err := connection.Write(response.GetBytesData()); err != nil {
			return
		}
	case USED_MEMORY:
		if len(message.Args) != 1 {
			return
		}

		userName := message.Args[0]
		log.Printf("Getting memory for " + userName)

		memory, err := GetUserUsedMemory(userName)
		if err != nil {
			log.Printf("Getting memory for " + userName + " error: " + err.Error())
			return
		}
		log.Printf("Getting memory for " + userName + " " + fmt.Sprint(memory))

		responseMessage := models.NewResponseMessage(0, []byte(fmt.Sprint(memory)))
		_ = connection.Write(responseMessage.GetBytesData())
	case CREATE:
		if len(message.Args) != 1 {
			return
		}

		fullPath := message.Args[0]

		if err := CreateDirectory(fullPath); err != nil {
			_ = SendResponseMessage(connection, 1, "")
			return
		}

		_ = SendResponseMessage(connection, 0, "")
	case RENAME:
		if len(message.Args) != 2 {
			_ = SendResponseMessage(connection, 1, "")
			return
		}

		fullPath := message.Args[0]
		newFullPath := message.Args[1]

		if err = RenameFileOrDirectory(fullPath, newFullPath); err != nil {
			_ = SendResponseMessage(connection, 1, "")
			return
		}

		_ = SendResponseMessage(connection, 0, "")
	case DELETE:
		if len(message.Args) != 1 {
			_ = SendResponseMessage(connection, 1, "")
			return
		}

		fullPath := message.Args[0]

		if err = DeleteFileOrDirectory(fullPath); err != nil {
			_ = SendResponseMessage(connection, 1, "")
			return
		}

		_ = SendResponseMessage(connection, 0, "")
	default:
		return
	}
}

func GetUserUsedMemory(username string) (int64, error) {
	entries, err := os.ReadDir(configurations.GetBaseFilesPath())
	if err != nil {
		return 0, err
	}

	for _, dir := range entries {
		if dir.Name() != username {
			continue
		}
		info, err := dir.Info()
		if err != nil {
			return 0, err
		}
		dirSize, err := dirSize(configurations.GetBaseFilesPath() + "/" + info.Name())
		if err != nil {
			return 0, err
		}
		return dirSize, nil
	}

	return 0, errors.New("username does not exist")
}

func GetFilesFromDirectory(path string) ([]models.File, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		return nil, errors.New("no directory with this path")
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var contents []models.File
	for _, file := range files {
		fileType, _ := getFileType(filepath.Join(path, file.Name()))
		fileDetails := models.File{Size: 0, Name: file.Name(), IsDir: file.IsDir(), Type: fileType}
		info, err := file.Info()
		if err == nil {
			fileDetails.Size = info.Size()
			fileDetails.Created = info.ModTime().Unix()
		}

		contents = append(contents, fileDetails)
	}

	return contents, nil
}

func dirSize(path string) (int64, error) {
	var dirSize int64 = 0

	readSize := func(path string, file os.FileInfo, err error) error {
		if file != nil && !file.IsDir() {
			dirSize += file.Size()
		}

		return nil
	}

	if err := filepath.Walk(path, readSize); err != nil {
		return 0, err
	}

	return dirSize, nil
}

func getFileType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil
	}
	defer file.Close()

	mimeType := mime.TypeByExtension(filePath)
	if mimeType == "" {
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return "", err
		}
		mimeType = http.DetectContentType(buffer[:n])
	}

	return mimeType, nil
}

func SendResponseMessage(mh *shared.MessageHandler, status byte, body string) error {
	message := models.NewResponseMessage(status, []byte(body))
	return mh.Write(message.GetBytesData())
}

func CreateDirectory(path string) error {
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFileOrDirectory(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func RenameFileOrDirectory(fullPath, newFullPath string) error {
	err := os.Rename(fullPath, newFullPath)
	if err != nil {
		return err
	}
	return nil
}
