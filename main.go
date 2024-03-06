package main

import (
	"NAS-Server-Web/commands"
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

func handleConnection(c net.Conn) {
	defer c.Close()
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	user := models.NewUser()
	connection := models.NewMessageHandler(c)
	for {
		rawMessage, err := connection.Read()
		if err != nil {
			return
		}

		message, err := models.NewMessage(rawMessage)
		if err != nil {
			continue
		}

		switch message.Command {
		case commands.UploadFile:
			commands.HandleUploadCommand(connection, &user, &message)
		case commands.DownloadFileOrDirectory:
			commands.HandleDownloadFileOrDirectory(connection, &user, &message)
			c.Close()
		case commands.CreateDirectory:
			commands.HandleCreateDirectoryCommand(connection, &user, &message)
		case commands.RemoveFileOrDirectory:
			commands.HandleRemoveFileOrDirectoryCommand(connection, &user, &message)
		case commands.RenameFileOrDirectory:
			commands.HandleRenameFileOrDirectoryCommand(connection, &user, &message)
		case commands.Login:
			commands.HandleLoginCommand(connection, &user, &message)
		case commands.ListFilesAndDirectories:
			commands.HandleListFilesAndDirectoriesCommand(connection, &user, &message)
		case commands.Info:
			commands.HandleInfoCommand(connection, &user, &message)
		default:
			continue
		}
	}
}

func GenX509KeyPair() (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(now.Unix()),
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, 1), // Valid for one day
		SubjectKeyId:          []byte{113, 117, 105, 99, 107, 115, 101, 114, 118, 101},
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template,
		priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}

func main() {
	println("Starting server...")
	service, err := services.NewConfigsService()
	if err != nil {
		panic(err)
	}

	cert, err := GenX509KeyPair()
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		Rand:         rand.Reader,
	}

	address := service.GetHost() + ":" + service.GetPort()
	listener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}

	log.Print("server: listening")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()

		log.Printf("server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		go handleConnection(conn)
	}
}
