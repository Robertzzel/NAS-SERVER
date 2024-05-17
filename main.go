package main

import (
	"NAS-Server-Web/commands"
	"NAS-Server-Web/configurations"
	"NAS-Server-Web/models"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"time"
)

func handleConnection(connection models.MessageHandler) {
	defer connection.Close()

	rawMessage, err := connection.Read()
	if err != nil {
		log.Print("Closed connection")
		return
	}

	message, err := models.NewMessage(rawMessage)
	if err != nil {
		log.Print("Bad message")
		return
	}

	if len(message.Args) < 2 {
		_ = connection.Write(append([]byte{1}, []byte("not enough arguments for login")...))
	}

	username := message.Args[0]
	password := message.Args[1]
	exists, err := commands.CheckUsernameAndPassword(username, password)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("error while checking username and password")...))
	}
	if !exists {
		_ = connection.Write(append([]byte{1}, []byte("invalid credentials")...))
	}
	// remove login credentials from message params
	message.Args = message.Args[2:]

	userDirectory := filepath.Join(configurations.BaseFilesBath, username)
	switch message.Command {
	case commands.UploadFile:
		log.Print("Started UploadFile with params:", message.Args, " ...")
		commands.UploadCommand(connection, &message, username, userDirectory)
		log.Print("Ended UploadFile with params:", message.Args, " ...")
	case commands.DownloadFileOrDirectory:
		log.Print("Started DownloadFileOrDirectory with params:", message.Args, " ...")
		commands.DownloadCommand(connection, &message, userDirectory)
		log.Print("Ended DownloadFileOrDirectory with params:", message.Args, " ...")
	case commands.CreateDirectory:
		log.Print("Started CreateDirectory with params:", message.Args, " ...")
		commands.CreateDirectoryCommand(connection, &message, userDirectory)
		log.Print("Ended CreateDirectory with params:", message.Args, " ...")
	case commands.RemoveFileOrDirectory:
		log.Print("Started RemoveFileOrDirectory with params:", message.Args, " ...")
		commands.RemoveCommand(connection, &message, userDirectory)
		log.Print("Ended RemoveFileOrDirectory with params:", message.Args, " ...")
	case commands.RenameFileOrDirectory:
		log.Print("Started RenameFileOrDirectory with params:", message.Args, " ...")
		commands.RenameCommand(connection, &message, userDirectory)
		log.Print("Ended RenameFileOrDirectory with params:", message.Args, " ...")
	case commands.Login:
		log.Print("Started Login with params:", message.Args, " ...")
		_ = connection.Write(append([]byte{0}, []byte("")...))
		log.Print("Ended Login with params:", message.Args, " ...")
	case commands.ListFilesAndDirectories:
		log.Print("Started ListFilesAndDirectories with params:", message.Args, " ...")
		commands.ListCommand(connection, &message, userDirectory)
		log.Print("Ended ListFilesAndDirectories with params:", message.Args, " ...")
	case commands.Info:
		log.Print("Started Info with params:", message.Args, " ...")
		commands.InfoCommand(connection, &message, username)
		log.Print("Ended Info with params:", message.Args, " ...")
	default:
		return
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
	log.Print("Starting server...")
	err := configurations.UpdateConfigurations()
	if err != nil {
		panic(err)
	}

	log.Print("Generating keys...")
	cert, err := GenX509KeyPair()
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		Rand:         rand.Reader,
	}

	address := configurations.Host + ":" + configurations.Port
	log.Print("Creating a TLS Server on ", address, "...")
	listener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}

	log.Print("Server listening...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()

		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if !ok {
			log.Printf("Connection does not have a valid TLS handshake from %s", conn.RemoteAddr())
			continue
		}

		state := tlscon.ConnectionState()
		for _, v := range state.PeerCertificates {
			log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
		}

		fmt.Printf("Serving %s\n", tlscon.RemoteAddr().String())
		messageHandler := models.NewMessageHandler(tlscon)
		go handleConnection(messageHandler)
	}
}
