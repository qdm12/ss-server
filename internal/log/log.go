package log

import (
	"fmt"
	"time"
)

type Level string

const (
	DebugLevel Level = "DEBUG"
	InfoLevel  Level = "INFO"
	ErrorLevel Level = "ERROR"
)

type Logger interface {
	Error(s string)
	Info(s string)
	Debug(s string)
}

func NewLogger(prefix string, level Level) Logger {
	return &logger{
		prefix:  prefix,
		level:   level,
		timeNow: time.Now,
	}
}

type logger struct {
	prefix  string
	level   Level
	timeNow func() time.Time
}

func (l *logger) log(level Level, message string) {
	fmt.Printf("%s [%s] %s%s\n", l.timeNow().Format(time.RFC3339), level, l.prefix, message)
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

func (l *logger) Error(s string) {
	l.log(ErrorLevel, s)
}
