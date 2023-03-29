package log

import "go.uber.org/zap"

var (
	Default *zap.Logger
)

// filed可以用zap.Int()等创建
func Debug(msg string, fields ...zap.Field) {
	Default.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func Debugf(msg string, value ...any) {
	Default.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func Info(msg string, fields ...zap.Field) {
	Default.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Infof(msg string, value ...any) {
	Default.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func Warn(msg string, fields ...zap.Field) {
	Default.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Warnf(msg string, fields ...any) {
	Default.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func Error(msg string, fields ...zap.Field) {
	Default.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Errorf(msg string, fields ...any) {
	Default.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func DevPanic(msg string, fields ...zap.Field) {
	Default.DPanic(msg, fields...)
}
func DevPanicf(msg string, fields ...any) {
	Default.Sugar().DPanicf(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Default.Panic(msg, fields...)
}
func Panicf(msg string, fields ...any) {
	Default.Sugar().Panicf(msg, fields...)
}

func Sync() error {
	return Default.Sync()
}
