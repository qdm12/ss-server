package tcpudp

import (
	"fmt"

	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
	"github.com/qdm12/ss-server/pkg/validation"
)

type Settings struct {
	// Listening address for the TCP and UDP servers.
	// It defaults to ":8388". It cannot be empty in the
	// internal state. Note it overrides the Address for both
	// the TCP and the UDP servers.
	Address string
	// LogAddresses to log addresses proxied for the TCP server.
	// It cannot be nil in the internal state.
	// Note it overrides the LogAddresses for both the TCP and the UDP servers.
	LogAddresses *bool
	// CipherName is the cipher to use for the TCP and UDP servers.
	// It defaults to chacha20-ietf-poly1305. It cannot be empty in the
	// internal state.
	CipherName string
	// Password for the TCP and UDP servers. It cannot be nil in the
	// internal state.
	Password *string

	// TCP can be used to set specific settings for the TCP server.
	TCP tcp.Settings
	// UDP can be used to set specific settings for the UDP server.
	UDP udp.Settings
}

// SetDefaults sets default values for all unset field
// in the settings.
func (s *Settings) SetDefaults() {
	if s.Address == "" {
		s.Address = ":8388"
	}

	if s.LogAddresses == nil {
		s.LogAddresses = new(bool)
	}

	if s.CipherName == "" {
		s.CipherName = core.Chacha20IetfPoly1305
	}

	if s.Password == nil {
		s.Password = new(string)
	}

	inheritedTCPSettings := s.toTCP()
	s.TCP.MergeWith(inheritedTCPSettings)
	s.TCP.SetDefaults()

	inheritedUDPSettings := s.toUDP()
	s.UDP.MergeWith(inheritedUDPSettings)
	s.UDP.SetDefaults()
}

// Copy returns a deep copy of the settings.
func (s Settings) Copy() (copied Settings) {
	copied.Address = s.Address
	if s.LogAddresses != nil {
		copied.LogAddresses = new(bool)
		*copied.LogAddresses = *s.LogAddresses
	}
	copied.CipherName = s.CipherName
	if s.Password != nil {
		copied.Password = new(string)
		*copied.Password = *s.Password
	}
	copied.TCP = s.TCP.Copy()
	copied.UDP = s.UDP.Copy()
	return copied
}

func (s Settings) toTCP() (settings tcp.Settings) {
	settings.Address = s.Address
	if s.LogAddresses != nil {
		settings.LogAddresses = new(bool)
		*settings.LogAddresses = *s.LogAddresses
	}
	settings.CipherName = s.CipherName
	if s.Password != nil {
		settings.Password = new(string)
		*settings.Password = *s.Password
	}
	return settings
}

func (s Settings) toUDP() (settings udp.Settings) {
	settings.Address = s.Address
	if s.LogAddresses != nil {
		settings.LogAddresses = new(bool)
		*settings.LogAddresses = *s.LogAddresses
	}
	settings.CipherName = s.CipherName
	if s.Password != nil {
		settings.Password = new(string)
		*settings.Password = *s.Password
	}
	return settings
}

// MergeWith sets unset fields of the receiving settings
// with field values from the other settings.
func (s *Settings) MergeWith(other Settings) {
	if s.Address == "" {
		s.Address = other.Address
	}

	if s.LogAddresses == nil && other.LogAddresses != nil {
		s.LogAddresses = new(bool)
		*s.LogAddresses = *other.LogAddresses
	}

	if s.CipherName == "" {
		s.CipherName = other.CipherName
	}

	if s.Password == nil && other.Password != nil {
		s.Password = new(string)
		*s.Password = *other.Password
	}

	s.TCP.MergeWith(other.TCP)
	s.UDP.MergeWith(other.UDP)
}

// OverrideWith sets any field of the receiving settings
// with the field value of any set field from the other settings.
func (s *Settings) OverrideWith(other Settings) {
	if other.Address != "" {
		s.Address = other.Address
	}

	if other.LogAddresses != nil {
		s.LogAddresses = new(bool)
		*s.LogAddresses = *other.LogAddresses
	}

	if other.CipherName != "" {
		s.CipherName = other.CipherName
	}

	if other.Password != nil {
		s.Password = new(string)
		*s.Password = *other.Password
	}

	s.TCP.OverrideWith(other.TCP)
	s.UDP.OverrideWith(other.UDP)
}

// Validate validates the settings are correct.
func (s *Settings) Validate() (err error) {
	err = validation.ValidateAddress(s.Address)
	if err != nil {
		return err
	}

	err = validation.ValidateCipher(s.CipherName)
	if err != nil {
		return err
	}

	err = s.TCP.Validate()
	if err != nil {
		return fmt.Errorf("failed validating TCP server settings: %w", err)
	}

	err = s.UDP.Validate()
	if err != nil {
		return fmt.Errorf("failed validating UDP server settings: %w", err)
	}

	return nil
}
