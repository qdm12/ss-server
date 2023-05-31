package tcp

type Logger interface {
	Debug(s string)
	Info(s string)
	Error(s string)
}
