package shared

import (
	"archive/zip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func GenX509KeyPair() (tls.Certificate, error) {
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
