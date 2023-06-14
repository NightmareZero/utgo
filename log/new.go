package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config LogConfig) (*zap.Logger, error) {
	setDefaultConfig(&config)

	// init logger encoderConfig
	var eConfig zapcore.Encoder
	if config.Dev {
		eConfig = getDevEncoder()
	} else {
		eConfig = getEncoder()
	}

	c := zapcore.NewTee(getZapCores(config, eConfig)...)
	if config.Caller {
		return zap.New(c, zap.AddCaller(), zap.AddCallerSkip(1)), nil
	}

	return zap.New(c), nil
}

func setDefaultConfig(config *LogConfig) error {
	// parse level
	var err error
	config.level, err = zapcore.ParseLevel(config.Level)
	if err != nil {
		return err
	}

	if config.CallerSkip == 0 {
		config.CallerSkip = 1
	}
	return nil
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
	Current = zlog{l}
	Normal = nlog{Current}
	return nil
}

func InitLog(level string) error {
	return InitWithConfig(LogConfig{
		Sync:      true,
		Level:     level,
		Caller:    true,
		Console:   true,
		NotToFile: false,
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
	Sync       bool   `json:"sync" yaml:"sync"`             // 启用同步模式
	Path       string `json:"path" yaml:"path"`             // 日志输出路径
	MergeError bool   `json:"mergeErr" yaml:"mergeErr"`     // 合并错误日志到常规日志
	ErrPath    string `json:"errPath" yaml:"errPath"`       // 错误日志输出路径
	Console    bool   `json:"console" yaml:"console"`       // 输出到控制台
	NotToFile  bool   `json:"file" yaml:"file"`             // 输出到文件
	Level      string `json:"level" yaml:"level"`           // 日志级别
	Dev        bool   `json:"dev" yaml:"dev"`               // 开发模式
	Caller     bool   `json:"caller" yaml:"caller"`         // 输出调用者信息
	CallerSkip int    `json:"callerSkip" yaml:"callerSkip"` // 跳过调用者代码层级(适用于封装层)

	level zapcore.Level
}
