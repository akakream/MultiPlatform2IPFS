package fs

import (
	"crypto/sha256"
	"encoding/json"
	"os"
)

func SaveJson(data any, filename string) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = os.WriteFile(filename, file, 0644)
}

func CreateDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}
}

func Sha256izeString(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return string(bs)
}