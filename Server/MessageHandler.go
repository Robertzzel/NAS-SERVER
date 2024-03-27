package Server

import (
	"archive/zip"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MessageHandler struct {
	conn io.ReadWriter
}

func NewMessageHandler(conn io.ReadWriter) *MessageHandler {
	return &MessageHandler{conn: conn}
}

func (mh *MessageHandler) Write(message []byte) error {
	size := uint64(len(message))

	err := binary.Write(mh.conn, binary.BigEndian, size)
	if err != nil {
		return err
	}

	_, err = mh.conn.Write(message)
	return err
}

func (mh *MessageHandler) Read() ([]byte, error) {
	var size uint64
	err := binary.Read(mh.conn, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}

	message := make([]byte, size)
	_, err = io.ReadFull(mh.conn, message)
	return message, err
}

func (mh *MessageHandler) ReadFile(writer io.Writer) error {
	var size uint64
	err := binary.Read(mh.conn, binary.BigEndian, &size)
	if err != nil {
		return err
	}

	_, err = io.CopyN(writer, mh.conn, int64(size))
	return err
}

func (mh *MessageHandler) SendFile(reader *os.File) error {
	_, err := io.Copy(mh.conn, reader)
	return err
}

func (mh *MessageHandler) SendDirectoryAsZip(inputDirectory, userRootDirectoryPath string) error {
	w := zip.NewWriter(mh.conn)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		inZipFile := strings.TrimPrefix(path, userRootDirectoryPath+"/")
		f, err := w.Create(inZipFile)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}
	return filepath.Walk(inputDirectory, walker)
}
