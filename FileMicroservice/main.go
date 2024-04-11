package main

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
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
	log.Printf("Starting...")
	cert, err := shared.GenX509KeyPair()
	if err != nil {
		panic(err)
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		Rand:         rand.Reader,
	}

	address := configurations.GetHost() + ":" + configurations.GetPort()
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

				address := configurations.GetHost() + ":" + fmt.Sprint(port)
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

				address := configurations.GetHost() + ":" + fmt.Sprint(port)
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
	// Attempt to connect to the port
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprint(port)), time.Second)
	if err != nil {
		// Port is closed or unreachable
		return false
	}
	defer conn.Close()

	// Port is open
	return true
}
