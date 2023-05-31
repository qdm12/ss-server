package settings

import (
	"fmt"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/log"
)

type Settings struct {
	CipherName string
	Password   *string
	Port       *uint16
	LogLevel   *log.Level
	Profiling  *bool
}

func (s *Settings) SetDefaults() {
	s.CipherName = gosettings.DefaultString(s.CipherName, "chacha20-ietf-poly1305")
	s.Password = gosettings.DefaultPointer(s.Password, "")
	const defaultPort = 8388
	s.Port = gosettings.DefaultPointer(s.Port, defaultPort)
	s.LogLevel = gosettings.DefaultPointer(s.LogLevel, log.LevelInfo)
	s.Profiling = gosettings.DefaultPointer(s.Profiling, false)
}

func (s *Settings) Validate() (err error) {
	err = validate.IsOneOf(s.CipherName, "chacha20-ietf-poly1305", "aes-256-gcm", "aes-128-gcm")
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	return nil
}
