package internal

type Logger interface {
	Error(template string, args ...interface{})
	Info(template string, args ...interface{})
}
