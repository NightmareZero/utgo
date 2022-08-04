package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLogger *zap.Logger
)

// filed可以用zap.Int()等创建
func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

// info
// 比上面的方法慢一倍，热点代码不建议用
func Debugf(msg string, value ...any) {
	defaultLogger.Sugar().Debugf(msg, value...)
}

// filed可以用zap.Int()等创建
func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Infof(msg string, value ...any) {
	defaultLogger.Sugar().Infof(msg, value...)
}

// warn
// filed可以用zap.Int()等创建
func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Warnf(msg string, fields ...any) {
	defaultLogger.Sugar().Warnf(msg, fields...)
}

// error
// filed可以用zap.Int()等创建
func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

// 比上面的方法慢一倍，热点代码不建议用
func Errorf(msg string, fields ...any) {
	defaultLogger.Sugar().Errorf(msg, fields...)
}

// 如果是开发模式，则抛出异常
func DevPanic(msg string, fields ...zap.Field) {
	defaultLogger.DPanic(msg, fields...)
}
func DevPanicf(msg string, fields ...any) {
	defaultLogger.Sugar().DPanicf(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	defaultLogger.Panic(msg, fields...)
}
func Panicf(msg string, fields ...any) {
	defaultLogger.Sugar().Panicf(msg, fields...)
}

func InitLog(sync bool) {
	// init logger encoderConfig
	eConfig := zap.NewDevelopmentConfig().EncoderConfig
	eConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// init logger output file
	logWriter := zapcore.AddSync(getWriter(sync, "./log", "log"))
	// init error logger output file
	errorLogWriter := zapcore.AddSync(getWriter(sync, "./err", "log"))

	tee := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig),
			logWriter, LogNormalLevel{}),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig),
			errorLogWriter, zap.ErrorLevel),
	)
	defaultLogger = zap.New(tee)
}

type LogNormalLevel struct {
}

func (e LogNormalLevel) Enabled(lvl zapcore.Level) bool {
	return lvl < zap.ErrorLevel
}