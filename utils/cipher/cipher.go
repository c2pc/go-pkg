package cipher

import (
	"encoding/hex"
	"io"
	"os"
	"strings"
)

type Cipher interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

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

var key2 = []byte("29HNL(<V}dv|GrzJ*7m5D<zK3i1!W:w@")

func getKey2() []byte {
	file, err := os.Open("key.key")
	if err != nil {
		return key2
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return key2
	}

	b2 := strings.TrimSpace(string(b))

	// Поддерживаем AES-128, AES-192 и AES-256
	if len(b2) >= 32 {
		b2 = b2[:32]
	} else if len(b2) >= 24 {
		b2 = b2[:24]
	} else if len(b2) >= 16 {
		b2 = b2[:16]
	} else {
		return key2
	}

	return []byte(b2)
}

func bsToString(bs []byte) string {
	str := string(bs)
	str = strings.Replace(str, "\u0000", "", -1)
	str = strings.Replace(str, "\x05", "", -1)
	return str
}

func hexToByte(src []byte) ([]byte, error) {
	dst := make([]byte, len(src))
	n, err := hex.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	return dst[:n], nil
}

func byteToHex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}

func Decrypt(ciphertext []byte) ([]byte, error) {
	aes, err := AESCipher.Decrypt(ciphertext)
	if err != nil {
		rc4, err := RC4Cipher.Decrypt(ciphertext)
		if err != nil {
			return nil, err
		}
		return rc4, nil
	}
	return aes, nil
}
