package tcpudp

import (
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/govalid/address"
	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
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
	s.Address = gosettings.DefaultString(s.Address, ":8388")
	s.LogAddresses = gosettings.DefaultPointer(s.LogAddresses, false)
	s.CipherName = gosettings.DefaultString(s.CipherName, core.Chacha20IetfPoly1305)
	s.Password = gosettings.DefaultPointer(s.Password, "")

	inheritedTCPSettings := s.toTCP()
	s.TCP = s.TCP.MergeWith(inheritedTCPSettings)
	s.TCP.SetDefaults()

	inheritedUDPSettings := s.toUDP()
	s.UDP = s.UDP.MergeWith(inheritedUDPSettings)
	s.UDP.SetDefaults()
}

// Copy returns a deep copy of the settings.
func (s Settings) Copy() (copied Settings) {
	copied.Address = s.Address
	copied.LogAddresses = gosettings.CopyPointer(s.LogAddresses)
	copied.CipherName = s.CipherName
	copied.Password = gosettings.CopyPointer(s.Password)
	copied.TCP = s.TCP.Copy()
	copied.UDP = s.UDP.Copy()
	return copied
}

func (s Settings) toTCP() (settings tcp.Settings) {
	settings.Address = s.Address
	settings.LogAddresses = gosettings.OverrideWithPointer(settings.LogAddresses, s.LogAddresses)
	settings.CipherName = s.CipherName
	settings.Password = gosettings.OverrideWithPointer(settings.Password, s.Password)
	return settings
}

func (s Settings) toUDP() (settings udp.Settings) {
	settings.Address = s.Address
	settings.LogAddresses = gosettings.OverrideWithPointer(settings.LogAddresses, s.LogAddresses)
	settings.CipherName = s.CipherName
	settings.Password = gosettings.OverrideWithPointer(settings.Password, s.Password)
	return settings
}

// MergeWith returns the merge result of the receiver settings with
// any unset fields set to the field of the other settings argument.
func (s *Settings) MergeWith(other Settings) (result Settings) {
	result.Address = gosettings.MergeWithString(s.Address, other.Address)
	result.LogAddresses = gosettings.MergeWithPointer(s.LogAddresses, other.LogAddresses)
	result.CipherName = gosettings.MergeWithString(s.CipherName, other.CipherName)
	result.Password = gosettings.MergeWithPointer(s.Password, other.Password)
	result.TCP = s.TCP.MergeWith(other.TCP)
	result.UDP = s.UDP.MergeWith(other.UDP)
	return result
}

// OverrideWith sets any field of the receiving settings
// with the field value of any set field from the other settings.
func (s *Settings) OverrideWith(other Settings) {
	s.Address = gosettings.OverrideWithString(s.Address, other.Address)
	s.LogAddresses = gosettings.OverrideWithPointer(s.LogAddresses, other.LogAddresses)
	s.CipherName = gosettings.OverrideWithString(s.CipherName, other.CipherName)
	s.Password = gosettings.OverrideWithPointer(s.Password, other.Password)
	s.TCP.OverrideWith(other.TCP)
	s.UDP.OverrideWith(other.UDP)
}

// Validate validates the settings are correct.
func (s *Settings) Validate() (err error) {
	err = address.Validate(s.Address, address.OptionListening(os.Getuid()))
	if err != nil {
		return fmt.Errorf("listening address: %w", err)
	}

	err = validate.IsOneOf(s.CipherName,
		core.AES128gcm, core.AES256gcm, core.Chacha20IetfPoly1305)
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	err = s.TCP.Validate()
	if err != nil {
		return fmt.Errorf("TCP server settings: %w", err)
	}

	err = s.UDP.Validate()
	if err != nil {
		return fmt.Errorf("UDP server settings: %w", err)
	}

	return nil
}
