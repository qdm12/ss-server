package core

import (
	"crypto/md5" //nolint: gosec
	"fmt"
	"strings"
)

// Derives a key from the password with a size depending on the cipher chosen
func deriveKey(password, cipherName string) (key []byte, err error) {
	keySize := 0
	switch strings.ToLower(cipherName) {
	case aes128gcm:
		keySize = 16
	case chacha20IetfPoly1305, aes256gcm:
		keySize = 32
	default:
		return nil, fmt.Errorf("cipher %q is not supported", cipherName)
	}
	return kdf(password, keySize)
}

// key derivation function from the original Shadowsocks spec based on md5
func kdf(password string, length int) (key []byte, err error) {
	var b, prev []byte
	h := md5.New() //nolint: go-sec
	for len(b) < length {
		if _, err := h.Write(prev); err != nil {
			return nil, err
		}
		if _, err := h.Write([]byte(password)); err != nil {
			return nil, err
		}
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:length], nil
}
