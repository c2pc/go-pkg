package cipher

import (
	"crypto/rc4"
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

var RC4Cipher Cipher

func init() {
	RC4Cipher = NewRC4(getKey())
}

type RC4 struct {
	key []byte
}

func NewRC4(key []byte) *RC4 {
	return &RC4{
		key: key,
	}
}

func (c *RC4) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := rc4.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(plaintext))
	block.XORKeyStream(out, plaintext)

	return byteToHex(out), nil
}

func (c *RC4) Decrypt(ciphertext []byte) ([]byte, error) {
	hx, err := hexToByte(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := rc4.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(hx))
	block.XORKeyStream(out, hx)

	if utf8.Valid(out) {
		return []byte(bsToString(out)), nil
	}

	return out, nil
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
