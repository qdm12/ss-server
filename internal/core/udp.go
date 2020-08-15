package core

import (
	"fmt"
	"net"
	"strings"

	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/shadowaead"
)

func NewUDPPacketCipher(name, password string, saltFilter filter.SaltFilter) (cipher UDPPacketCipher, err error) {
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
	return &udpPacketCipher{
		aead:       aead,
		saltFilter: saltFilter,
	}, nil
}

type UDPPacketCipher interface {
	Shadow(connection net.PacketConn) net.PacketConn
}

type udpPacketCipher struct {
	aead       shadowaead.AEADCipher
	saltFilter filter.SaltFilter
}

func (c *udpPacketCipher) Shadow(connection net.PacketConn) net.PacketConn {
	return shadowaead.NewPacketConn(connection, c.aead, c.saltFilter)
}
