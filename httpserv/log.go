package httpserv

type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warn(v ...any)
	Error(v ...any)
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
}
