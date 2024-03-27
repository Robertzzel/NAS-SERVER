package Server

import (
	"NAS-Server-Web/shared/configurations"
	models2 "NAS-Server-Web/shared/models"
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

func StartServer() {
	log.Print("Starting server...")
	err := configurations.UpdateConfigurations()
	if err != nil {
		panic(err)
	}

	log.Print("Generating keys...")
	cert, err := genX509KeyPair()
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		Rand:         rand.Reader,
	}

	address := configurations.GetHost() + ":" + configurations.GetPort()
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

func handleConnection(c net.Conn) {
	defer c.Close()
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	user := models2.NewUser()
	connection := NewMessageHandler(c)
	for {
		rawMessage, err := connection.Read()
		if err != nil {
			log.Print("Closed connection with ", c.RemoteAddr())
			return
		}

		message, err := models2.NewMessage(rawMessage)
		if err != nil {
			log.Print("Bad message structure from ", c.RemoteAddr())
			continue
		}

		switch message.Command {
		case UploadFile:
			log.Print("Started UploadFile with params:", message.Args, " ...")
			HandleUploadCommand(connection, &message)
			log.Print("Ended UploadFile with params:", message.Args, " ...")
		case DownloadFileOrDirectory:
			log.Print("Started DownloadFileOrDirectory with params:", message.Args, " ...")
			HandleDownloadFileOrDirectory(connection, &user, &message)
			log.Print("Closing connection...")
			_ = c.Close()
			log.Print("Ended DownloadFileOrDirectory with params:", message.Args, " ...")
		case CreateDirectory:
			log.Print("Started CreateDirectory with params:", message.Args, " ...")
			HandleCreateDirectoryCommand(connection, &user, &message)
			log.Print("Ended CreateDirectory with params:", message.Args, " ...")
		case RemoveFileOrDirectory:
			log.Print("Started RemoveFileOrDirectory with params:", message.Args, " ...")
			HandleRemoveFileOrDirectoryCommand(connection, &user, &message)
			log.Print("Ended RemoveFileOrDirectory with params:", message.Args, " ...")
		case RenameFileOrDirectory:
			log.Print("Started RenameFileOrDirectory with params:", message.Args, " ...")
			HandleRenameFileOrDirectoryCommand(connection, &user, &message)
			log.Print("Ended RenameFileOrDirectory with params:", message.Args, " ...")
		case Login:
			log.Print("Started Login with params:", message.Args, " ...")
			HandleLoginCommand(connection, &user, &message)
			log.Print("Ended Login with params:", message.Args, " ...")
		case ListFilesAndDirectories:
			log.Print("Started ListFilesAndDirectories with params:", message.Args, " ...")
			HandleListFilesAndDirectoriesCommand(connection, &user, &message)
			log.Print("Ended ListFilesAndDirectories with params:", message.Args, " ...")
		case Info:
			log.Print("Started Info with params:", message.Args, " ...")
			HandleInfoCommand(connection, &user, &message)
			log.Print("Ended Info with params:", message.Args, " ...")
		default:
			continue
		}
	}
}

func genX509KeyPair() (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(now.Unix()),
		NotBefore:             now,
		NotAfter:              now.AddDate(1, 0, 0),
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
