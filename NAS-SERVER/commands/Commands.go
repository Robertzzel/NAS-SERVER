package commands

import (
	"NAS-Server-Web/models"
	"NAS-Server-Web/services"
	"errors"
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

func HandleUploadCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if len(message.Args) != 4 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := checkUsernameAndPassword(username, password)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte(err.Error())).Data)
		return
	}
	if !exists {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	filename := message.Args[2]
	size, err := strconv.Atoi(message.Args[3])
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid size")).Data)
		return
	}

	remainingMemory, err := services.GetUserRemainingMemory(username)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if remainingMemory < int64(size) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("no memory for the upload")).Data)
		return
	}

	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	config, err := services.NewConfigsService()
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
	userRootDirectory := filepath.Join(config.BaseFilesBath, username)
	filename = path.Join(userRootDirectory, filename)
	file, err := os.Create(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
	defer file.Close()

	_ = connection.Write(models.NewMessageForClient(0, []byte("go on")).Data)

	if err = connection.ReadFile(file); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)
}

func HandleDownloadFileOrDirectory(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if len(message.Args) != 3 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := checkUsernameAndPassword(username, password)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte(err.Error())).Data)
		return
	}
	if !exists {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	filename := message.Args[2]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	config, err := services.NewConfigsService()
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	userRootDirectory := filepath.Join(config.BaseFilesBath, username)
	filename = path.Join(userRootDirectory, filename)
	stat, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if stat.IsDir() {
		_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)

		err := connection.SendDirectoryAsZip(filename, user.UserRootDirectory)
		if err != nil {
			_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
			return
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
			return
		}
		defer file.Close()

		_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)

		if err = connection.SendFile(file); err != nil {
			_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
			return
		}
	}
}

func HandleCreateDirectoryCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	if err := os.Mkdir(filename, os.ModePerm); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)
}

func HandleRemoveFileOrDirectoryCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	_, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
	if err := os.RemoveAll(filename); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("")).Data)
}

func HandleRenameFileOrDirectoryCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 2 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	filename := message.Args[0]
	newFilename := message.Args[1]
	if !IsPathSafe(filename) && !IsPathSafe(newFilename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	filename = path.Join(user.UserRootDirectory, filename)
	newFilename = path.Join(user.UserRootDirectory, newFilename)

	if err := os.Rename(filename, newFilename); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)
}

func HandleLoginCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	config, err := services.NewConfigsService()
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if len(message.Args) != 2 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	username := message.Args[0]
	password := message.Args[1]

	exists, err := checkUsernameAndPassword(username, password)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte(err.Error())).Data)
		return
	}
	if exists {
		user.IsAuthenticated = true
		user.Name = username
		user.UserRootDirectory = filepath.Join(config.BaseFilesBath, username)
	} else {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid username or password")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)
}

func HandleListFilesAndDirectoriesCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	directoryPath := message.Args[0]
	if !IsPathSafe(directoryPath) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	directoryPath = path.Join(user.UserRootDirectory, directoryPath)
	directory, err := services.GetFilesFromDirectory(directoryPath)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	resultMessage := ""
	for file := range directory {
		resultMessage += directory[file].Name + "\n" + strconv.FormatInt(directory[file].Size, 10) + "\n" + strconv.FormatBool(directory[file].IsDir) + "\n" + directory[file].Type + "\n" + strconv.FormatInt(directory[file].Created, 10) + "\x1c"
	}
	if len(resultMessage) > 0 {
		resultMessage = resultMessage[:len(resultMessage)-1]
	}

	if err := connection.Write(models.NewMessageForClient(0, []byte(resultMessage)).Data); err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}
}

func HandleInfoCommand(connection *models.MessageHandler, user *models.User, message *models.MessageForServer) {
	if !user.IsAuthenticated {
		_ = connection.Write(models.NewMessageForClient(1, []byte("user is not authenticated")).Data)
		return
	}

	if len(message.Args) != 0 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	remainingMemory, err := services.GetUserRemainingMemory(user.Name)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	_ = connection.Write(models.NewMessageForClient(0, []byte(strconv.FormatInt(remainingMemory, 10))).Data)
}

func IsPathSafe(path string) bool {
	return !strings.Contains(path, "../")
}

func checkUsernameAndPassword(name, password string) (bool, error) {
	db, err := services.NewDatabaseService()
	if err != nil {
		return false, errors.New("internal error")
	}

	return db.UsernameAndPasswordExists(name, password)
}
