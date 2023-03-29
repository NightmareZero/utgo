package log

import "go.uber.org/zap"

type zlog struct {
	log *zap.Logger
}

// filed可以用zap.Int()等创建
func (l zlog) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Debugf(msg string, value ...any) {
	l.log.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func (l zlog) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Infof(msg string, value ...any) {
	l.log.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func (l zlog) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Warnf(msg string, fields ...any) {
	l.log.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func (l zlog) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func (l zlog) Errorf(msg string, fields ...any) {
	l.log.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func (l zlog) DevPanic(msg string, fields ...zap.Field) {
	l.log.DPanic(msg, fields...)
}
func (l zlog) DevPanicf(msg string, fields ...any) {
	l.log.Sugar().DPanicf(msg, fields...)
}

func (l zlog) Panic(msg string, fields ...zap.Field) {
	l.log.Panic(msg, fields...)
}
func (l zlog) Panicf(msg string, fields ...any) {
	l.log.Sugar().Panicf(msg, fields...)
}

func (l zlog) Sync() error {
	return l.log.Sync()
}

// filed可以用zap.Int()等创建
func Debug(msg string, fields ...zap.Field) {
	Current.log.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func Debugf(msg string, value ...any) {
	Current.log.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func Info(msg string, fields ...zap.Field) {
	Current.log.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Infof(msg string, value ...any) {
	Current.log.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func Warn(msg string, fields ...zap.Field) {
	Current.log.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Warnf(msg string, fields ...any) {
	Current.log.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func Error(msg string, fields ...zap.Field) {
	Current.log.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Errorf(msg string, fields ...any) {
	Current.log.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func DevPanic(msg string, fields ...zap.Field) {
	Current.log.DPanic(msg, fields...)
}
func DevPanicf(msg string, fields ...any) {
	Current.log.Sugar().DPanicf(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Current.log.Panic(msg, fields...)
}
func Panicf(msg string, fields ...any) {
	Current.log.Sugar().Panicf(msg, fields...)
}

func Sync() error {
	return Current.log.Sync()
}
