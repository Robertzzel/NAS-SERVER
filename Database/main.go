package main

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"strconv"
)

func main() {
	log.Printf("Starting...")
	db, err := sql.Open("sqlite3", configurations.GetDatabasePath())
	if err != nil {
		panic(err)
	}

	err = MigrateDatabase(db)
	if err != nil {
		panic(err)
	}

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

		go handleConnection(tlscon, db)
	}

}

func handleConnection(c net.Conn, db *sql.DB) {
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
		case 0: // check username and password
			if len(message.Args) != 2 {
				continue
			}
			exists, err := UsernameAndPasswordExists(db, message.Args[0], message.Args[1])
			if err != nil {
				continue
			}
			var response []byte
			if exists {
				response = []byte{1}
			} else {
				response = []byte{0}
			}
			responseMessage := models.NewResponseMessage(0, response)
			_ = connection.Write(responseMessage.GetBytesData())
		case 1: // get user allocated memory
			if len(message.Args) != 1 {
				continue
			}
			memory, err := GetUserAllocatedMemory(db, message.Args[0])
			if err != nil {
				continue
			}
			responseMessage := models.NewResponseMessage(0, []byte(fmt.Sprint(memory)))
			_ = connection.Write(responseMessage.GetBytesData())
		case 2: // add user
			if len(message.Args) != 3 {
				continue
			}

			memory, err := strconv.Atoi(message.Args[2])
			if err != nil {
				continue
			}

			if err = AddUser(db, message.Args[0], message.Args[1], memory); err != nil {
				continue
			}

			responseMessage := models.NewResponseMessage(0, []byte(fmt.Sprint("Success")))

			_ = connection.Write(responseMessage.GetBytesData())
		default:
			continue
		}
	}
}

func UsernameAndPasswordExists(db *sql.DB, username, password string) (bool, error) {
	var cnt int
	err := db.QueryRow(`select count(*) from User where Name = ? and Password = ? LIMIT 1`, username, hash(password)).Scan(&cnt)
	if err != nil {
		return false, errors.New("database problem")
	}
	return cnt != 0, nil
}

func GetUserAllocatedMemory(db *sql.DB, username string) (uint64, error) {
	var memory uint64
	err := db.QueryRow(`select AllocatedMemory from User where Name = ? LIMIT 1`, username).Scan(&memory)
	if err != nil {
		return 0, err
	}
	return memory, nil
}

func AddUser(db *sql.DB, username, password string, memory int) error {
	_, err := db.Exec(`INSERT INTO User (Name, Password, AllocatedMemory) VALUES (?, ?, ?)`, username, hash(password), memory)
	return err
}

func MigrateDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS User(
    	Id integer PRIMARY KEY,
		Name varchar(255) UNIQUE NOT NULL,
		Password varchar(255) NOT NULL,
    	AllocatedMemory integer NOT NULL
    )`)
	if err != nil {
		return err
	}
	return nil
}

func hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}

func uint64ToBytes(n uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(n))
	return bytes
}

func bytesToInt(bytes []byte) int {
	return int(binary.BigEndian.Uint64(bytes))
}
