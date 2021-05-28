package log

import (
	"os"

	"github.com/qdm12/ss-server/internal/log"
)

type Logger interface {
	Debug(s string)
	Info(s string)
	Error(s string)
}

func New() Logger {
	return log.New(log.InfoLevel, os.Stdout)
}
