package log

var (
	Current zlog
	Normal  Logger
)

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

type nlog struct {
	zlog
}

// filed可以用zap.Int()等创建
func (l nlog) Debug(fields ...any) {
	l.Log.Sugar().Debug(fields...)
}

// filed可以用zap.Int()等创建
func (l nlog) Info(fields ...any) {
	l.Log.Sugar().Info(fields...)
}

// warn
// filed可以用zap.Int()等创建
func (l nlog) Warn(fields ...any) {
	l.Log.Sugar().Warn(fields...)
}

// error
// filed可以用zap.Int()等创建
func (l nlog) Error(fields ...any) {
	l.Log.Sugar().Error(fields...)
}
