package cipher

import (
	"io"
	"os"
	"strings"
)

var key = []byte("asdfdsgf")

func getKey() []byte {
	file, err := os.Open("key.key")
	if err != nil {
		return key
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return key
	}

	b2 := strings.TrimSpace(string(b))

	return []byte(b2)
}

type Cipher interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}
