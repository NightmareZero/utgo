package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config LogConfig) (*zap.Logger, error) {
	// parse level
	var err error
	config.level, err = zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	// init logger encoderConfig
	var eConfig zapcore.Encoder
	if config.Dev {
		eConfig = getDevEncoder()
	} else {
		eConfig = getEncoder()
	}

	c := zapcore.NewTee(getZapCores(config, eConfig)...)
	if config.Caller {
		return zap.New(c, zap.AddCaller()), nil
	}

	return zap.New(c), nil
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

func InitWithConfig(config LogConfig) error {
	l, err := NewLogger(config)
	if err != nil {
		return err
	}
	defaultLogger = l
	return nil
}

func InitLog(level string) error {
	return InitWithConfig(LogConfig{
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
	Level         string
	level         zapcore.Level
	Dev           bool
	Caller        bool
}
