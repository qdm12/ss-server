package settings

import (
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/govalid/address"
	"github.com/qdm12/log"
)

type Settings struct {
	CipherName string
	Password   *string
	Address    *string
	LogLevel   *log.Level
	Profiling  *bool
}

func (s *Settings) SetDefaults() {
	s.CipherName = gosettings.DefaultString(s.CipherName, "chacha20-ietf-poly1305")
	s.Password = gosettings.DefaultPointer(s.Password, "")
	s.Address = gosettings.DefaultPointer(s.Address, ":8388")
	s.LogLevel = gosettings.DefaultPointer(s.LogLevel, log.LevelInfo)
	s.Profiling = gosettings.DefaultPointer(s.Profiling, false)
}

func (s *Settings) Validate() (err error) {
	err = validate.IsOneOf(s.CipherName, "chacha20-ietf-poly1305", "aes-256-gcm", "aes-128-gcm")
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	err = address.Validate(*s.Address, address.OptionListening(os.Geteuid()))
	if err != nil {
		return fmt.Errorf("listening address: %w", err)
	}

	return nil
}
