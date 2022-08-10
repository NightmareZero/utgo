package log

import (
	"os"

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

	cores := getZapCores(config, eConfig)

	defaultLogger = zap.New(zapcore.NewTee(cores...))
}

func getZapCores(config LogConfig, eConfig zap.Config) (ret []zapcore.Core) {
	fixedPath := uos.FixPathEndSlash(common.If(len(config.Path) > 0, config.Path, "./log/"))

	// 初始化 log 文件输出
	logWriter := zapcore.AddSync(getFileWriter(config.Sync, fixedPath, "log"))
	ret = append(ret, zapcore.NewCore(
		zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
		logWriter, LogNormalLevel{config.Level, config.MergeErrorLog}))

	// 如果未开启日志合并
	if !config.MergeErrorLog {
		fixedErrPath := uos.FixPathEndSlash(common.If(len(config.ErrPath) > 0, config.ErrPath, "./log/"))
		errorLogWriter := zapcore.AddSync(getFileWriter(config.Sync, fixedErrPath, "err"))
		ret = append(ret, zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
			errorLogWriter, zap.ErrorLevel))
	}

	// 如果开启屏幕输出
	if config.Console {
		ret = append(ret, zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
			os.Stdout, LogNormalLevel{config.Level, false}))
		ret = append(ret, zapcore.NewCore(
			zapcore.NewConsoleEncoder(eConfig.EncoderConfig),
			os.Stderr, zap.ErrorLevel))
	}

	return
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
