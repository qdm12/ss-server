package core

import (
	"fmt"
	"net"
	"strings"

	"github.com/qdm12/ss-server/internal/shadowaead"
)

func NewUDPPacketCipher(name, password string, saltFilter SaltFilter) (
	cipher *UDPPacketCipher, err error) {
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
		return nil, fmt.Errorf("%w: for UDP: %s", ErrCipherNotSupported, name)
	}
	return &UDPPacketCipher{
		aead:       aead,
		saltFilter: saltFilter,
	}, nil
}

type UDPPacketCipher struct {
	aead       *shadowaead.AEADCipherAdapter
	saltFilter SaltFilter
}

func (c *UDPPacketCipher) Shadow(connection net.PacketConn) net.PacketConn {
	return shadowaead.NewPacketConn(connection, c.aead, c.saltFilter)
}
