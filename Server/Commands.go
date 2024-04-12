package main

import (
	"NAS-Server-Web/shared"
	"NAS-Server-Web/shared/Services"
	"NAS-Server-Web/shared/configurations"
	"NAS-Server-Web/shared/models"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	UploadFile              = 0
	DownloadFileOrDirectory = 1
	CreateDirectory         = 2
	RemoveFileOrDirectory   = 3
	RenameFileOrDirectory   = 4
	Login                   = 5
	ListFilesAndDirectories = 6
	Info                    = 7
)

func HandleUploadCommand(userService *Services.DatabaseService, connection *shared.MessageHandler, message *models.RequestMessage) {
	if len(message.Args) != 4 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := userService.CheckUsernameAndPassword(username, password)
	if err != nil {
		_ = SendResponseMessage(connection, 1, err.Error())
		return
	}
	if !exists {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	filename := message.Args[2]
	size, err := strconv.Atoi(message.Args[3])
	if err != nil {
		_ = SendResponseMessage(connection, 1, "invalid size")
		return
	}

	usedMemory, err := Services.GetUserUsedMemory(username)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	allocatedMemory, err := userService.GetUserAllocatedMemory(username)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	remainingMemory := int64(allocatedMemory) - usedMemory
	if remainingMemory < int64(size) {
		_ = SendResponseMessage(connection, 1, "no memory for the upload")
		return
	}

	if !IsPathSafe(filename) {
		_ = SendResponseMessage(connection, 1, "bad path")
		return
	}

	userRootDirectory := filepath.Join(configurations.GetBaseFilesPath(), username)
	filename = path.Join(userRootDirectory, filename)

	// TODO REPLACE
	file, err := os.Create(filename)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}
	defer file.Close()

	_ = SendResponseMessage(connection, 0, "go on")

	if err = connection.ReadFile(file); err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}
	//
	_ = SendResponseMessage(connection, 0, "")
}

func HandleDownloadFileOrDirectory(userService *Services.DatabaseService, connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if len(message.Args) != 3 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := userService.CheckUsernameAndPassword(username, password)
	if err != nil {
		_ = SendResponseMessage(connection, 1, err.Error())
		return
	}
	if !exists {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	filename := message.Args[2]
	if !IsPathSafe(filename) {
		_ = SendResponseMessage(connection, 1, "bad path")
		return
	}

	userRootDirectory := filepath.Join(configurations.GetBaseFilesPath(), username)
	filename = path.Join(userRootDirectory, filename)
	// TODO REPLACE
	stat, err := os.Stat(filename)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	if stat.IsDir() {
		_ = SendResponseMessage(connection, 0, "success")

		err := connection.SendDirectoryAsZip(filename, user.UserRootDirectory)
		if err != nil {
			_ = SendResponseMessage(connection, 1, "internal error")
			return
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			_ = SendResponseMessage(connection, 1, "internal error")
			return
		}
		defer file.Close()

		_ = SendResponseMessage(connection, 0, "")

		if err = connection.SendFile(file); err != nil {
			_ = SendResponseMessage(connection, 1, "internal error")
			return
		}
	}
	//
}

func HandleCreateDirectoryCommand(connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if !user.IsAuthenticated {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	if len(message.Args) != 1 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = SendResponseMessage(connection, 1, "bad path")
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	if err := Services.CreateDirectory(filename); err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	_ = SendResponseMessage(connection, 0, "")
}

func HandleRemoveFileOrDirectoryCommand(connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if !user.IsAuthenticated {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	if len(message.Args) != 1 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = SendResponseMessage(connection, 1, "bad path")
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)

	err := Services.DeleteFileOrDirectory(filename)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	_ = SendResponseMessage(connection, 0, "")
}

func HandleRenameFileOrDirectoryCommand(connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if !user.IsAuthenticated {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	if len(message.Args) != 2 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	filename := message.Args[0]
	newFilename := message.Args[1]
	if !IsPathSafe(filename) && !IsPathSafe(newFilename) {
		_ = SendResponseMessage(connection, 1, "bad path")
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	newFilename = path.Join(user.UserRootDirectory, newFilename)

	if err := Services.RenameFileOrDirectory(filename, newFilename); err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}
	_ = SendResponseMessage(connection, 0, "success")
}

func HandleLoginCommand(userService *Services.DatabaseService, connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if len(message.Args) != 2 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := userService.CheckUsernameAndPassword(username, password)
	if err != nil {
		log.Println("Username")
		_ = SendResponseMessage(connection, 1, err.Error())
		return
	}
	if exists {
		user.IsAuthenticated = true
		user.Name = username
		user.UserRootDirectory = filepath.Join(configurations.GetBaseFilesPath(), username)
	} else {
		_ = SendResponseMessage(connection, 1, "invalid username or password")
		return
	}

	_ = SendResponseMessage(connection, 0, "success")
}

func HandleListFilesAndDirectoriesCommand(connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if !user.IsAuthenticated {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	if len(message.Args) != 1 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	directoryPath := message.Args[0]
	if !IsPathSafe(directoryPath) {
		_ = SendResponseMessage(connection, 1, "bad path")
		return
	}

	directoryPath = path.Join(user.UserRootDirectory, directoryPath)
	directory, err := Services.GetFilesFromDirectory(directoryPath)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}
	if err := SendResponseMessage(connection, 0, directory); err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}
}

func HandleInfoCommand(userService *Services.DatabaseService, connection *shared.MessageHandler, user *models.User, message *models.RequestMessage) {
	if !user.IsAuthenticated {
		_ = SendResponseMessage(connection, 1, "user is not authenticated")
		return
	}

	if len(message.Args) != 0 {
		_ = SendResponseMessage(connection, 1, "invalid number of arguments")
		return
	}

	usedMemory, err := Services.GetUserUsedMemory(user.Name)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	allocatedMemory, err := userService.GetUserAllocatedMemory(user.Name)
	if err != nil {
		_ = SendResponseMessage(connection, 1, "internal error")
		return
	}

	remainingMemory := int64(allocatedMemory) - usedMemory

	_ = SendResponseMessage(connection, 0, strconv.FormatInt(remainingMemory, 10))
}

func IsPathSafe(path string) bool {
	return !strings.Contains(path, "../")
}

func SendResponseMessage(mh *shared.MessageHandler, status byte, body string) error {
	message := models.NewResponseMessage(status, []byte(body))
	return mh.Write(message.GetBytesData())
}
