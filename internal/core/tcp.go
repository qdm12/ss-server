package core

import (
	"fmt"
	"net"
	"strings"

	"github.com/qdm12/ss-server/internal/shadowaead"
)

func NewTCPStreamCipher(name, password string, saltFilter SaltFilter) (
	cipher *TCPStreamCipher, err error) {
	key, err := deriveKey(password, name)
	if err != nil {
		return nil, err
	}
	var aead *shadowaead.AEADCipherAdapter
	switch strings.ToLower(name) {
	case Chacha20IetfPoly1305:
		aead = shadowaead.Chacha20Poly1305(key)
	case AES128gcm, AES256gcm:
		aead = shadowaead.AESGCM(key)
	default:
		return nil, fmt.Errorf("%w: for TCP: %s", ErrCipherNotSupported, name)
	}
	return &TCPStreamCipher{
		aead:       aead,
		saltFilter: saltFilter,
	}, nil
}

type TCPStreamCipher struct {
	aead       *shadowaead.AEADCipherAdapter
	saltFilter SaltFilter
}

func (c *TCPStreamCipher) Shadow(connection net.Conn) net.Conn {
	return shadowaead.NewConn(connection, c.aead, c.saltFilter)
}
