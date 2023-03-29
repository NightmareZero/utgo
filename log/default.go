package log

import "go.uber.org/zap"

var (
	Current _log
)

type _log struct {
	zlog *zap.Logger
}

// filed可以用zap.Int()等创建
func (l _log) Debug(msg string, fields ...zap.Field) {
	l.zlog.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func (l _log) Debugf(msg string, value ...any) {
	l.zlog.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func (l _log) Info(msg string, fields ...zap.Field) {
	l.zlog.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l _log) Infof(msg string, value ...any) {
	l.zlog.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func (l _log) Warn(msg string, fields ...zap.Field) {
	l.zlog.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l _log) Warnf(msg string, fields ...any) {
	l.zlog.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func (l _log) Error(msg string, fields ...zap.Field) {
	l.zlog.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l _log) Errorf(msg string, fields ...any) {
	l.zlog.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func (l _log) DevPanic(msg string, fields ...zap.Field) {
	l.zlog.DPanic(msg, fields...)
}
func (l _log) DevPanicf(msg string, fields ...any) {
	l.zlog.Sugar().DPanicf(msg, fields...)
}

func (l _log) Panic(msg string, fields ...zap.Field) {
	l.zlog.Panic(msg, fields...)
}
func (l _log) Panicf(msg string, fields ...any) {
	l.zlog.Sugar().Panicf(msg, fields...)
}

func (l _log) Sync() error {
	return l.zlog.Sync()
}

// filed可以用zap.Int()等创建
func Debug(msg string, fields ...zap.Field) {
	Current.zlog.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func Debugf(msg string, value ...any) {
	Current.zlog.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func Info(msg string, fields ...zap.Field) {
	Current.zlog.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Infof(msg string, value ...any) {
	Current.zlog.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func Warn(msg string, fields ...zap.Field) {
	Current.zlog.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Warnf(msg string, fields ...any) {
	Current.zlog.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func Error(msg string, fields ...zap.Field) {
	Current.zlog.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Errorf(msg string, fields ...any) {
	Current.zlog.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func DevPanic(msg string, fields ...zap.Field) {
	Current.zlog.DPanic(msg, fields...)
}
func DevPanicf(msg string, fields ...any) {
	Current.zlog.Sugar().DPanicf(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Current.zlog.Panic(msg, fields...)
}
func Panicf(msg string, fields ...any) {
	Current.zlog.Sugar().Panicf(msg, fields...)
}

func Sync() error {
	return Current.zlog.Sync()
}
