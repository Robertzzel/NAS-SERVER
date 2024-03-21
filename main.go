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
			log.Print("Closed connection with ", c.RemoteAddr())
			return
		}

		message, err := models.NewMessage(rawMessage)
		if err != nil {
			log.Print("Bad message structure from ", c.RemoteAddr())
			continue
		}

		switch message.Command {
		case commands.UploadFile:
			log.Print("Started UploadFile with params:", message.Args, " ...")
			commands.HandleUploadCommand(connection, &message)
			log.Print("Ended UploadFile with params:", message.Args, " ...")
		case commands.DownloadFileOrDirectory:
			log.Print("Started DownloadFileOrDirectory with params:", message.Args, " ...")
			commands.HandleDownloadFileOrDirectory(connection, &user, &message)
			log.Print("Closing connection...")
			_ = c.Close()
			log.Print("Ended DownloadFileOrDirectory with params:", message.Args, " ...")
		case commands.CreateDirectory:
			log.Print("Started CreateDirectory with params:", message.Args, " ...")
			commands.HandleCreateDirectoryCommand(connection, &user, &message)
			log.Print("Ended CreateDirectory with params:", message.Args, " ...")
		case commands.RemoveFileOrDirectory:
			log.Print("Started RemoveFileOrDirectory with params:", message.Args, " ...")
			commands.HandleRemoveFileOrDirectoryCommand(connection, &user, &message)
			log.Print("Ended RemoveFileOrDirectory with params:", message.Args, " ...")
		case commands.RenameFileOrDirectory:
			log.Print("Started RenameFileOrDirectory with params:", message.Args, " ...")
			commands.HandleRenameFileOrDirectoryCommand(connection, &user, &message)
			log.Print("Ended RenameFileOrDirectory with params:", message.Args, " ...")
		case commands.Login:
			log.Print("Started Login with params:", message.Args, " ...")
			commands.HandleLoginCommand(connection, &user, &message)
			log.Print("Ended Login with params:", message.Args, " ...")
		case commands.ListFilesAndDirectories:
			log.Print("Started ListFilesAndDirectories with params:", message.Args, " ...")
			commands.HandleListFilesAndDirectoriesCommand(connection, &user, &message)
			log.Print("Ended ListFilesAndDirectories with params:", message.Args, " ...")
		case commands.Info:
			log.Print("Started Info with params:", message.Args, " ...")
			commands.HandleInfoCommand(connection, &user, &message)
			log.Print("Ended Info with params:", message.Args, " ...")
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
	log.Print("Starting server...")
	service, err := services.NewConfigsService()
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

	address := service.Host + ":" + service.Port
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
		go handleConnection(conn)
	}
}
