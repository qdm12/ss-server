package core //nolint:dupl

import (
	"fmt"
	"net"
	"strings"

	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/shadowaead"
)

func NewTCPStreamCipher(name, password string, saltFilter filter.SaltFilter) (
	cipher *TCPStreamCipher, err error) {
	key, err := deriveKey(password, name)
	if err != nil {
		return nil, err
	}
	var aead shadowaead.AEADCipher
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

var _ ConnShadower = (*TCPStreamCipher)(nil)

type ConnShadower interface {
	Shadow(connection net.Conn) net.Conn
}

type TCPStreamCipher struct {
	aead       shadowaead.AEADCipher
	saltFilter filter.SaltFilter
}

func (c *TCPStreamCipher) Shadow(connection net.Conn) net.Conn {
	return shadowaead.NewConn(connection, c.aead, c.saltFilter)
}
