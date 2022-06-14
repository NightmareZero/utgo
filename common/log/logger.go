package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// 如果是开发模式，则抛出异常
func DevPanic(msg string, fields ...zap.Field) {
	logger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}

func init() {
	// init logger encoderConfig
	eConfig := zap.NewDevelopmentConfig().EncoderConfig
	eConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// init logger output file
	logWriter := zapcore.AddSync(getWriter("./log", "log"))
	// init error logger output file
	errorLogWriter := zapcore.AddSync(getWriter("./err", "log"))

	tee := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig),
			logWriter, LogNormalLevel{}),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig),
			errorLogWriter, zap.ErrorLevel),
	)
	logger = zap.New(tee)

}

type LogNormalLevel struct {
}

func (e LogNormalLevel) Enabled(lvl zapcore.Level) bool {
	return lvl < zap.ErrorLevel
}
