package log

import (
	"github.com/NightmareZero/nzgoutil/common"
	"github.com/NightmareZero/nzgoutil/uos"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitWithConfig(config LogConfig) {
	// init logger encoderConfig
	var eConfig zap.Config
	if config.Dev {
		eConfig = zap.NewDevelopmentConfig()
	} else {
		eConfig = zap.NewProductionConfig()
	}
	eConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var core zapcore.Core

	if !config.MergeErrorLog {
		core = newMultiCoreLogger(config, eConfig)
	} else {
		core = newSingleCoreLogger(config, eConfig)
	}

	defaultLogger = zap.New(core)
}

// 生成 info,error 合并 的日志文件写入
func newSingleCoreLogger(config LogConfig, eConfig zap.Config) zapcore.Core {
	fixedPath := uos.FixPathEndSlash(common.If(len(config.Path) > 0, config.Path, "./log/"))

	// init logger output file
	logWriter := zapcore.AddSync(getWriter(config.Sync, fixedPath, "log"))

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
		logWriter, config.Level)
}

// 生成 info,error 分离 的日志文件写入
func newMultiCoreLogger(config LogConfig, eConfig zap.Config) zapcore.Core {
	fixedPath := uos.FixPathEndSlash(common.If(len(config.Path) > 0, config.Path, "./log/"))

	// init logger output file
	logWriter := zapcore.AddSync(getWriter(config.Sync, fixedPath, "log"))
	fixedErrPath := uos.FixPathEndSlash(common.If(len(config.ErrPath) > 0, config.ErrPath, "./log/"))
	// init error logger output file
	errorLogWriter := zapcore.AddSync(getWriter(config.Sync, fixedErrPath, "err"))
	return zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
			logWriter, LogNormalLevel{config.Level, config.MergeErrorLog}),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
			errorLogWriter, zap.ErrorLevel),
	)
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
	return e.Level <= lvl && (!e.MergeErrorLog || lvl < zap.ErrorLevel)
}

type LogConfig struct {
	Sync          bool
	Path          string
	MergeErrorLog bool
	ErrPath       string
	Level         zapcore.Level
	Dev           bool
}
