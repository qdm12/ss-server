package udp

import (
	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/pkg/validation"
)

type Settings struct {
	// Listening address for the TCP server.
	// It defaults to ":8388".
	// It cannot be empty in the internal state.
	Address string
	// LogAddresses can be set to true to log
	// addresses proxied though the TCP server.
	// It defaults to false.
	// It cannot be nil in the internal state.
	LogAddresses *bool
	// CipherName is the cipher to use for the TCP server.
	// It defaults to "chacha20-ietf-poly1305".
	// It cannot be empty in the internal state.
	CipherName string
	// Password for the TCP server.
	// It defaults to the empty string.
	// It cannot be nil in the internal state.
	Password *string
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
	return copied
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
}

func (s *Settings) Validate() (err error) {
	err = validation.ValidateAddress(s.Address)
	if err != nil {
		return err
	}

	err = validation.ValidateCipher(s.CipherName)
	if err != nil {
		return err
	}

	return nil
}
