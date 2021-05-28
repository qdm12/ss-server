package env

import (
	"strings"

	"github.com/qdm12/ss-server/internal/log"
)

type Reader interface {
	CipherName() (cipherName string)
	Password() (password string)
	Port() (port string)
	LogLevel() (logLevel log.Level)
	Profiling() (profiling bool)
}

type reader struct {
	envKV map[string]string
}

func NewReader(environ []string) Reader {
	kv := make(map[string]string, len(environ))
	for _, s := range environ {
		parts := strings.Split(s, "=")
		key := parts[0]
		value := parts[1]
		kv[key] = value
	}

	return &reader{
		envKV: kv,
	}
}

func (r *reader) CipherName() (cipherName string) {
	cipherName = r.envKV["CIPHER"]
	if cipherName == "" {
		const defaultCipherName = "chacha20-ietf-poly1305"
		cipherName = defaultCipherName
	}
	return cipherName
}

func (r *reader) Password() (password string) {
	password = r.envKV["PASSWORD"]
	if password == "" {
		const defaultPassword = "password"
		password = defaultPassword
	}
	return password
}

func (r *reader) Port() (port string) {
	port = r.envKV["PORT"]
	if port == "" {
		const defaultPort = "8388"
		port = defaultPort
	}
	return port
}

func (r *reader) LogLevel() (logLevel log.Level) {
	logLevel = log.Level(r.envKV["LOG_LEVEL"])
	if logLevel == "" {
		const defaultLogLevel = log.InfoLevel
		logLevel = defaultLogLevel
	}
	return logLevel
}

func (r *reader) Profiling() (profiling bool) {
	return strings.EqualFold(r.envKV["PROFILING"], "on")
}
