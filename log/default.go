package log

import "go.uber.org/zap"

type zlog struct {
	Log *zap.Logger
}

// filed可以用zap.Int()等创建
func (l zlog) Debug(msg string, fields ...zap.Field) {
	l.Log.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Debugf(msg string, value ...any) {
	l.Log.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func (l zlog) Info(msg string, fields ...zap.Field) {
	l.Log.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Infof(msg string, value ...any) {
	l.Log.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func (l zlog) Warn(msg string, fields ...zap.Field) {
	l.Log.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Warnf(msg string, fields ...any) {
	l.Log.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func (l zlog) Error(msg string, fields ...zap.Field) {
	l.Log.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Errorf(msg string, fields ...any) {
	l.Log.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func (l zlog) DevPanic(msg string, fields ...zap.Field) {
	l.Log.DPanic(msg, fields...)
}
func (l zlog) DevPanicf(msg string, fields ...any) {
	l.Log.Sugar().DPanicf(msg, fields...)
}

func (l zlog) Panic(msg string, fields ...zap.Field) {
	l.Log.Panic(msg, fields...)
}
func (l zlog) Panicf(msg string, fields ...any) {
	l.Log.Sugar().Panicf(msg, fields...)
}

func (l zlog) Sync() error {
	return l.Log.Sync()
}

// filed可以用zap.Int()等创建
func Debug(msg string, fields ...zap.Field) {
	Current.Log.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func Debugf(msg string, value ...any) {
	Current.Log.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func Info(msg string, fields ...zap.Field) {
	Current.Log.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Infof(msg string, value ...any) {
	Current.Log.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func Warn(msg string, fields ...zap.Field) {
	Current.Log.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Warnf(msg string, fields ...any) {
	Current.Log.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func Error(msg string, fields ...zap.Field) {
	Current.Log.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Errorf(msg string, fields ...any) {
	Current.Log.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func DevPanic(msg string, fields ...zap.Field) {
	Current.Log.DPanic(msg, fields...)
}
func DevPanicf(msg string, fields ...any) {
	Current.Log.Sugar().DPanicf(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Current.Log.Panic(msg, fields...)
}
func Panicf(msg string, fields ...any) {
	Current.Log.Sugar().Panicf(msg, fields...)
}

func Sync() error {
	return Current.Log.Sync()
}
