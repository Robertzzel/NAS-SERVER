package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func DownloadCommand(connection models.MessageHandler, message *models.MessageForServer, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(models.NewMessageForClient(1, []byte("invalid number of arguments")).Data)
		return
	}

	// get the requested path for the file
	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(models.NewMessageForClient(1, []byte("bad path")).Data)
		return
	}

	//prepend the user root directory
	filename = path.Join(clientFileDirectory, filename)
	stat, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(models.NewMessageForClient(1, []byte("internal error")).Data)
		return
	}

	if stat.IsDir() {
		_ = connection.Write(models.NewMessageForClient(0, []byte("success")).Data)

		err := connection.SendDirectoryAsZip(filename, clientFileDirectory)
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
