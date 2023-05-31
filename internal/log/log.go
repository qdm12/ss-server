package log

import (
	"io"
	"log"
)

type Level string

const (
	DebugLevel Level = "DEBUG"
	InfoLevel  Level = "INFO"
	WarnLevel  Level = "WARN"
	ErrorLevel Level = "ERROR"
)

func New(level Level, w io.Writer) *SimpleLogger {
	flags := log.Ldate | log.Ltime | log.Lshortfile
	logImpl := log.New(w, "", flags)
	return &SimpleLogger{
		level:  level,
		logger: logImpl,
	}
}

type SimpleLogger struct {
	level  Level
	logger *log.Logger
}

func (l *SimpleLogger) log(level Level, message string) {
	const callDepth = 3
	_ = l.logger.Output(callDepth, "["+string(level)+"] "+message)
}

func (l *SimpleLogger) Debug(s string) {
	if l.level == DebugLevel {
		l.log(DebugLevel, s)
	}
}

func (l *SimpleLogger) Info(s string) {
	if l.level == DebugLevel || l.level == InfoLevel {
		l.log(InfoLevel, s)
	}
}

func (l *SimpleLogger) Warn(s string) {
	switch l.level {
	case DebugLevel, InfoLevel, WarnLevel:
		l.log(WarnLevel, s)
	case ErrorLevel:
	}
}

func (l *SimpleLogger) Error(s string) {
	l.log(ErrorLevel, s)
}
