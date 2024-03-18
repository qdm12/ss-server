package config

import (
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
	"github.com/qdm12/log"
)

type Settings struct {
	CipherName string
	Password   *string
	Address    *string
	LogLevel   string
	Profiling  *bool
}

func (s *Settings) SetDefaults() {
	s.CipherName = gosettings.DefaultComparable(s.CipherName, "chacha20-ietf-poly1305")
	s.Password = gosettings.DefaultPointer(s.Password, "")
	s.Address = gosettings.DefaultPointer(s.Address, ":8388")
	s.LogLevel = gosettings.DefaultComparable(s.LogLevel, "info")
	s.Profiling = gosettings.DefaultPointer(s.Profiling, false)
}

func (s *Settings) Validate() (err error) {
	err = validate.IsOneOf(s.CipherName, "chacha20-ietf-poly1305", "aes-256-gcm", "aes-128-gcm")
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	err = validate.ListeningAddress(*s.Address, os.Geteuid())
	if err != nil {
		return fmt.Errorf("listening address: %w", err)
	}

	_, err = log.ParseLevel(s.LogLevel)
	if err != nil {
		return fmt.Errorf("log level: %w", err)
	}

	return nil
}

func (s *Settings) ToLinesNode() *gotree.Node {
	node := gotree.New("Settings summary:")
	node.Appendf("Listening address: " + *s.Address)
	node.Appendf("Cipher name: " + s.CipherName)
	node.Appendf("Password: " + gosettings.ObfuscateKey(*s.Password))
	node.Appendf("Log level: " + s.LogLevel)
	node.Appendf("Profiling: " + gosettings.BoolToYesNo(s.Profiling))
	return node
}

func (s Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) Read(reader *reader.Reader) (err error) {
	s.CipherName = reader.String("CIPHER")
	s.Password = reader.Get("PASSWORD")
	s.Address = reader.Get("LISTENING_ADDRESS")
	s.LogLevel = reader.String("LOG_LEVEL")
	s.Profiling, err = reader.BoolPtr("PROFILING")
	if err != nil {
		return err
	}
	return nil
}
