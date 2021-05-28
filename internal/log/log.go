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

type Logger interface {
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
}

func New(level Level, w io.Writer) Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile
	logImpl := log.New(w, "", flags)
	return &logger{
		level:  level,
		logger: logImpl,
	}
}

type logger struct {
	level  Level
	logger *log.Logger
}

func (l *logger) log(level Level, message string) {
	const callDepth = 3
	_ = l.logger.Output(callDepth, "["+string(level)+"] "+message)
}

func (l *logger) Debug(s string) {
	if l.level == DebugLevel {
		l.log(DebugLevel, s)
	}
}

func (l *logger) Info(s string) {
	if l.level == DebugLevel || l.level == InfoLevel {
		l.log(InfoLevel, s)
	}
}

func (l *logger) Warn(s string) {
	switch l.level {
	case DebugLevel, InfoLevel, WarnLevel:
		l.log(WarnLevel, s)
	case ErrorLevel:
	}
}

func (l *logger) Error(s string) {
	l.log(ErrorLevel, s)
}
