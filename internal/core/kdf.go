package core

import (
	"crypto/md5" //nolint:gosec
	"errors"
	"fmt"
	"strings"
)

var ErrCipherNotSupported = errors.New("cipher is not supported")

// Derives a key from the password with a size depending on the cipher chosen.
func deriveKey(password, cipherName string) (key []byte, err error) {
	var keySize int
	switch strings.ToLower(cipherName) {
	case AES128gcm:
		keySize = 16
	case Chacha20IetfPoly1305, AES256gcm:
		keySize = 32
	default:
		return nil, fmt.Errorf("%w: %s", ErrCipherNotSupported, cipherName)
	}
	return kdf(password, keySize)
}

// key derivation function from the original Shadowsocks spec based on md5.
func kdf(password string, length int) (key []byte, err error) {
	var b, prev []byte
	hasher := md5.New() //nolint:gosec
	for len(b) < length {
		if _, err := hasher.Write(prev); err != nil {
			return nil, err
		}
		if _, err := hasher.Write([]byte(password)); err != nil {
			return nil, err
		}
		b = hasher.Sum(b)
		prev = b[len(b)-hasher.Size():]
		hasher.Reset()
	}
	return b[:length], nil
}
