package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config LogConfig) *zap.Logger {
	// init logger encoderConfig
	var eConfig zap.Config
	if config.Dev {
		eConfig = zap.NewDevelopmentConfig()
	} else {
		eConfig = zap.NewProductionConfig()
	}
	eConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cores := getZapCores(config, eConfig)

	return zap.New(zapcore.NewTee(cores...))
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
