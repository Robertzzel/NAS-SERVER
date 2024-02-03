package services

import (
	"NAS-Server-Web/models"
	_ "encoding/json"
	"errors"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func UploadFile(username, filename string, reader io.Reader, size int64) error {
	remainingMemory, err := GetUserRemainingMemory(username)
	if err != nil {
		return errors.New("internal error")
	}

	if remainingMemory < size {
		return errors.New("no memory for the upload")
	}

	if !IsPathSafe(filename) {
		return errors.New("bad path")
	}

	create, err := os.Create(filename)
	if err != nil {
		return errors.New("internal error")
	}

	_, err = io.Copy(create, reader)
	if err != nil {
		return errors.New("internal error")
	}

	return nil
}

func RenameFile(oldPath, newPath string) error {
	if !IsPathSafe(oldPath) || !IsPathSafe(newPath) {
		return errors.New("bad parameters")
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return errors.New("internal error")
	}

	return nil
}

func DirSize(path string) (int64, error) {
	var dirSize int64 = 0

	readSize := func(path string, file os.FileInfo, err error) error {
		if file != nil && !file.IsDir() {
			dirSize += file.Size()
		}

		return nil
	}

	if err := filepath.Walk(path, readSize); err != nil {
		return 0, err
	}

	return dirSize, nil
}

func IsPathSafe(path string) bool {
	return !strings.Contains(path, "../")
}

func GetFilesFromDirectory(path string) ([]models.FileDetails, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		return nil, errors.New("no directory with this path")
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var contents []models.FileDetails
	for _, file := range files {
		fileType, _ := GetFileType(filepath.Join(path, file.Name()))
		fileDetails := models.FileDetails{Size: 0, Name: file.Name(), IsDir: file.IsDir(), Type: fileType}
		info, err := file.Info()
		if err == nil {
			fileDetails.Size = info.Size()
			fileDetails.Created = info.ModTime().Unix()
		}

		contents = append(contents, fileDetails)
	}

	return contents, nil
}

func GetUserRemainingMemory(username string) (int64, error) {
	used, err := GetUserUsedMemory(username)
	if err != nil {
		return 0, err
	}

	db, err := NewDatabaseService()
	if err != nil {
		return 0, err
	}

	allocated, err := db.GetUserAllocatedMemory(username)
	if err != nil {
		return 0, err
	}

	return int64(allocated) - used, nil
}

func GetUserUsedMemory(username string) (int64, error) {
	configs, err := NewConfigsService()
	if err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(configs.GetBaseFilesPath())
	if err != nil {
		return 0, err
	}

	for _, dir := range entries {
		if dir.Name() != username {
			continue
		}
		info, err := dir.Info()
		if err != nil {
			return 0, err
		}
		dirSize, err := DirSize(configs.GetBaseFilesPath() + "/" + info.Name())
		if err != nil {
			return 0, err
		}
		return dirSize, nil
	}

	return 0, errors.New("username does not exist")
}

func GetFileType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil
	}
	defer file.Close()

	mimeType := mime.TypeByExtension(filePath)
	if mimeType == "" {
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return "", err
		}
		mimeType = http.DetectContentType(buffer[:n])
	}

	return mimeType, nil
}
