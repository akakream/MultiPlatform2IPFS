package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

func SaveJson(data any, filename string) error {
	file, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, file, 0644)
}

func WriteBytesToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}
	return nil
}

func CreateDirs(paths []string) error {
	for _, path := range paths {
		if err := CreateDir(path); err != nil {
			return err
		}
	}
	return nil
}

func Sha256izeString(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return string(bs)
}

func Sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
