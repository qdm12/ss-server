package udp

import "github.com/qdm12/ss-server/internal/core"

type Settings struct {
	// Listening address for the UDP server.
	// It defaults to ":8388".
	Address string
	// LogAddresses to log addresses proxied for the TCP server.
	LogAddresses bool
	// CipherName is the cipher to use for the UDP server.
	// It defaults to chacha20-ietf-poly1305.
	CipherName string
	// Password for the UDP server.
	Password string
}

func (s *Settings) setDefaults() {
	if s.Address == "" {
		s.Address = ":8388"
	}

	if s.CipherName == "" {
		s.CipherName = core.Chacha20IetfPoly1305
	}
}
