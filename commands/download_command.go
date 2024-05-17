package commands

import (
	"NAS-Server-Web/models"
	"os"
	"path"
)

func DownloadCommand(connection models.MessageHandler, message *models.Message, clientFileDirectory string) {
	if len(message.Args) != 1 {
		_ = connection.Write(append([]byte{1}, []byte("invalid number of arguments")...))
		return
	}

	// get the requested path for the file
	filename := message.Args[0]
	if !IsPathSafe(filename) {
		_ = connection.Write(append([]byte{1}, []byte("bad path")...))
		return
	}

	//prepend the user root directory
	filename = path.Join(clientFileDirectory, filename)
	stat, err := os.Stat(filename)
	if err != nil {
		_ = connection.Write(append([]byte{1}, []byte("internal error")...))
		return
	}

	if stat.IsDir() {
		_ = connection.Write(append([]byte{0}, []byte("success")...))

		err := connection.SendDirectoryAsZip(filename, clientFileDirectory)
		if err != nil {
			_ = connection.Write(append([]byte{1}, []byte("internal error")...))
			return
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			_ = connection.Write(append([]byte{1}, []byte("internal error")...))
			return
		}
		defer file.Close()

		_ = connection.Write(append([]byte{0}, []byte("")...))

		if err = connection.SendFile(file); err != nil {
			_ = connection.Write(append([]byte{1}, []byte("internal error")...))
			return
		}
	}
}
