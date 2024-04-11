package main

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	DOWNLOAD    = 0
	UPLOAD      = 1
	LIST        = 2
	USED_MEMORY = 3
)

func main() {
	cert, err := shared.GenX509KeyPair()
	if err != nil {
		panic(err)
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		Rand:         rand.Reader,
	}

	err = configurations.UpdateConfigurations()
	if err != nil {
		return
	}
	address := configurations.GetFilesHost() + ":" + configurations.GetFilesPort()
	log.Printf("Starting at " + address + "...")
	listener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if !ok {
			log.Printf("Connection does not have a valid TLS handshake from %s", conn.RemoteAddr())
			continue
		}

		go handleConnection(tlscon)
	}

}

func handleConnection(c net.Conn) {
	defer c.Close()
	connection := shared.NewMessageHandler(c)

	for {
		rawMessage, err := connection.Read()
		if err != nil {
			log.Print("Closed connection with ", c.RemoteAddr())
			return
		}

		message, err := models.NewRequestMessageFromBytes(rawMessage)
		if err != nil {
			log.Print("Bad message structure from ", c.RemoteAddr())
			continue
		}

		switch message.Command {
		case DOWNLOAD:
			if len(message.Args) != 1 {
				continue
			}

			filePath := message.Args[0]

			port := 10000
			for {
				if IsPortOpen(port) {
					break
				}
				port++
			}

			go func(filePath string, port int) {
				cert, err := shared.GenX509KeyPair()
				if err != nil {
					return
				}

				config := tls.Config{
					Certificates: []tls.Certificate{cert},
					MinVersion:   tls.VersionTLS13,
					Rand:         rand.Reader,
				}

				address := configurations.GetFilesHost() + ":" + fmt.Sprint(port)
				listener, err := tls.Listen("tcp", address, &config)
				if err != nil {
					return
				}

				_, err = listener.Accept()
			}(filePath, port)

			response := models.NewResponseMessage(0, []byte(fmt.Sprint(port)))
			if err := connection.Write(response.GetBytesData()); err != nil {
				continue
			}
		case UPLOAD:
			if len(message.Args) != 1 {
				continue
			}

			filePath := message.Args[0]

			port := 10000
			for {
				if IsPortOpen(port) {
					break
				}
				port++
			}

			go func(filePath string, port int) {
				cert, err := shared.GenX509KeyPair()
				if err != nil {
					return
				}

				config := tls.Config{
					Certificates: []tls.Certificate{cert},
					MinVersion:   tls.VersionTLS13,
					Rand:         rand.Reader,
				}

				address := configurations.GetFilesHost() + ":" + fmt.Sprint(port)
				listener, err := tls.Listen("tcp", address, &config)
				if err != nil {
					return
				}

				_, err = listener.Accept()
			}(filePath, port)

			response := models.NewResponseMessage(0, []byte(fmt.Sprint(port)))
			if err := connection.Write(response.GetBytesData()); err != nil {
				continue
			}
		case LIST:
			if len(message.Args) != 1 {
				continue
			}

			directoryPath := message.Args[0]

			directory, err := GetFilesFromDirectory(directoryPath)
			if err != nil {
				continue
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
				continue
			}
		case USED_MEMORY:
			if len(message.Args) != 1 {
				continue
			}

			userName := message.Args[0]

			memory, err := GetUserUsedMemory(userName)
			if err != nil {
				continue
			}

			responseMessage := models.NewResponseMessage(0, []byte(fmt.Sprint(memory)))
			_ = connection.Write(responseMessage.GetBytesData())
		default:
			continue
		}
	}
}

func IsPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprint(port)), time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func GetUserUsedMemory(username string) (int64, error) {
	entries, err := os.ReadDir(configurations.GetDatabasePath())
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
		dirSize, err := dirSize(configurations.GetDatabasePath() + "/" + info.Name())
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
