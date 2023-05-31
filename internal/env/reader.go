package env

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qdm12/log"
)

type Reader struct {
	envKV map[string]string
}

func NewReader(environ []string) *Reader {
	keyValue := make(map[string]string, len(environ))
	for _, s := range environ {
		parts := strings.Split(s, "=")
		key := parts[0]
		value := parts[1]
		keyValue[key] = value
	}

	return &Reader{
		envKV: keyValue,
	}
}

func (r *Reader) CipherName() (cipherName string) {
	cipherName = r.envKV["CIPHER"]
	if cipherName == "" {
		const defaultCipherName = "chacha20-ietf-poly1305"
		cipherName = defaultCipherName
	}
	return cipherName
}

func (r *Reader) Password() (password string) {
	password = r.envKV["PASSWORD"]
	if password == "" {
		const defaultPassword = "password"
		password = defaultPassword
	}
	return password
}

func (r *Reader) Port() (port string) {
	port = r.envKV["PORT"]
	if port == "" {
		const defaultPort = "8388"
		port = defaultPort
	}
	return port
}

var (
	ErrLogLevelUnknown = errors.New("log level is unknown")
)

func (r *Reader) LogLevel() (logLevel log.Level, err error) {
	value := r.envKV["LOG_LEVEL"]
	if value == "" {
		return log.LevelInfo, nil
	}

	validLevels := []log.Level{log.LevelDebug, log.LevelInfo,
		log.LevelWarn, log.LevelError}
	for _, validLevel := range validLevels {
		if strings.EqualFold(value, validLevel.String()) {
			return validLevel, nil
		}
	}

	return 0, fmt.Errorf("%w: %s", ErrLogLevelUnknown, value)
}

func (r *Reader) Profiling() (profiling bool) {
	return strings.EqualFold(r.envKV["PROFILING"], "on")
}
