package tcpudp

import (
	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
)

type Settings struct {
	// Listening address for the TCP and UDP servers.
	// It defaults to ":8388".
	Address string
	// LogAddresses to log addresses proxied for the TCP server.
	LogAddresses bool
	// CipherName is the cipher to use for the TCP and UDP servers.
	// It defaults to chacha20-ietf-poly1305.
	CipherName string
	// Password for the TCP and UDP servers.
	Password string

	// TCP can be used to set specific settings for the TCP server.
	TCP tcp.Settings
	// UDP can be used to set specific settings for the UDP server.
	UDP udp.Settings
}

func (s *Settings) setDefaults() {
	if s.Address == "" {
		s.Address = ":8388"
	}

	if s.CipherName == "" {
		s.CipherName = core.Chacha20IetfPoly1305
	}
}

func (s *Settings) propagateToTCPAndUDP() {
	s.setDefaults() // ensure top level defaults are set

	if s.TCP.Address == "" {
		s.TCP.Address = s.Address
	}
	if s.UDP.Address == "" {
		s.UDP.Address = s.Address
	}

	if !s.TCP.LogAddresses {
		s.TCP.LogAddresses = s.LogAddresses
	}
	if !s.UDP.LogAddresses {
		s.UDP.LogAddresses = s.LogAddresses
	}

	if s.TCP.CipherName == "" {
		s.TCP.CipherName = s.CipherName
	}
	if s.UDP.CipherName == "" {
		s.UDP.CipherName = s.CipherName
	}

	if s.TCP.Password == "" {
		s.TCP.Password = s.Password
	}
	if s.UDP.Password == "" {
		s.UDP.Password = s.Password
	}
}
