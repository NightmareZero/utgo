package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config LogConfig) *zap.Logger {
	// init logger encoderConfig
	var eConfig zapcore.Encoder
	if config.Dev {
		eConfig = getDevEncoder()
	} else {
		eConfig = getEncoder()
	}

	cores := getZapCores(config, eConfig)

	return zap.New(zapcore.NewTee(cores...))
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getDevEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func InitWithConfig(config LogConfig) {
	defaultLogger = NewLogger(config)
}

func InitLog(level zapcore.Level) {
	InitWithConfig(LogConfig{
		Sync:  true,
		Level: level,
	})
}

type LogNormalLevel struct {
	Level         zapcore.Level
	MergeErrorLog bool
}

func (e LogNormalLevel) Enabled(lvl zapcore.Level) bool {
	return e.Level <= lvl && (e.MergeErrorLog || lvl < zap.ErrorLevel)
}

type LogConfig struct {
	Sync          bool
	Path          string
	MergeErrorLog bool
	ErrPath       string
	Console       bool
	Level         zapcore.Level
	Dev           bool
}
