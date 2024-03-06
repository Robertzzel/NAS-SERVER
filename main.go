package main

import (
	"NAS-Server-Web/commands"
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"fmt"
	"net"
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

func main() {
	println("Starting server...")
	service, err := services.NewConfigsService()
	if err != nil {
		panic(err)
	}

	address := service.GetHost() + ":" + service.GetPort()

	l, err := net.Listen("tcp4", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	println("Server started on " + address)
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
