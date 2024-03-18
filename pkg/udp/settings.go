package udp

import (
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/govalid/address"
	"github.com/qdm12/ss-server/internal/core"
)

type Settings struct {
	// Listening address for the TCP server.
	// It defaults to ":8388".
	// It cannot be nil in the internal state.
	Address *string
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
	s.Address = gosettings.DefaultPointer(s.Address, ":8388")
	s.LogAddresses = gosettings.DefaultPointer(s.LogAddresses, false)
	s.CipherName = gosettings.DefaultComparable(s.CipherName, core.Chacha20IetfPoly1305)
	s.Password = gosettings.DefaultPointer(s.Password, "")
}

// Copy returns a deep copy of the settings.
func (s Settings) Copy() (copied Settings) {
	copied.Address = s.Address
	copied.LogAddresses = gosettings.CopyPointer(s.LogAddresses)
	copied.CipherName = s.CipherName
	copied.Password = gosettings.CopyPointer(s.Password)
	return copied
}

// OverrideWith sets any field of the receiving settings
// with the field value of any set field from the other settings.
func (s *Settings) OverrideWith(other Settings) {
	s.Address = gosettings.OverrideWithPointer(s.Address, other.Address)
	s.LogAddresses = gosettings.OverrideWithPointer(s.LogAddresses, other.LogAddresses)
	s.CipherName = gosettings.OverrideWithComparable(s.CipherName, other.CipherName)
	s.Password = gosettings.OverrideWithPointer(s.Password, other.Password)
}

func (s *Settings) Validate() (err error) {
	err = address.Validate(*s.Address, address.OptionListening(os.Getuid()))
	if err != nil {
		return fmt.Errorf("listening address: %w", err)
	}

	err = validate.IsOneOf(s.CipherName,
		core.AES128gcm, core.AES256gcm, core.Chacha20IetfPoly1305)
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	return nil
}
