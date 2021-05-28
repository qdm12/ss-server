package core //nolint:dupl

import (
	"fmt"
	"net"
	"strings"

	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/shadowaead"
)

func NewTCPStreamCipher(name, password string, saltFilter filter.SaltFilter) (cipher TCPStreamCipher, err error) {
	key, err := deriveKey(password, name)
	if err != nil {
		return nil, err
	}
	var aead shadowaead.AEADCipher
	switch strings.ToLower(name) {
	case chacha20IetfPoly1305:
		aead = shadowaead.Chacha20Poly1305(key)
	case aes128gcm, aes256gcm:
		aead = shadowaead.AESGCM(key)
	default:
		return nil, fmt.Errorf("cipher %q is not supported", name)
	}
	return &tcpStreamCipher{
		aead:       aead,
		saltFilter: saltFilter,
	}, nil
}

type TCPStreamCipher interface {
	Shadow(connection net.Conn) net.Conn
}

type tcpStreamCipher struct {
	aead       shadowaead.AEADCipher
	saltFilter filter.SaltFilter
}

func (c *tcpStreamCipher) Shadow(connection net.Conn) net.Conn {
	return shadowaead.NewConn(connection, c.aead, c.saltFilter)
}
