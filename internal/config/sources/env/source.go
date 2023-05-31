package env

import (
	"github.com/qdm12/gosettings/sources/env"
	"github.com/qdm12/log"
	"github.com/qdm12/ss-server/internal/config/settings"
)

type Source struct{}

func New() *Source {
	return &Source{}
}

func (s *Source) Read() (settings settings.Settings, err error) {
	settings.CipherName = env.String("CIPHER_NAME")
	settings.Password = env.Get("PASSWORD", env.ForceLowercase(false))
	settings.Address = env.Get("LISTENING_ADDRESS")
	settings.LogLevel, err = s.readLogLevel()
	if err != nil {
		return settings, err
	}
	settings.Profiling, err = env.BoolPtr("PROFILING")
	if err != nil {
		return settings, err
	}
	return settings, nil
}

func (s *Source) readLogLevel() (logLevel *log.Level, err error) {
	value := env.Get("LOG_LEVEL")
	if value == nil {
		return nil, nil //nolint:nilnil
	}

	level, err := log.ParseLevel(*value)
	if err != nil {
		return nil, err
	}
	return &level, nil
}
