package log

import (
	"io"
	"os"
	"time"

	"github.com/NightmareZero/nzgoutil/uos"
	"github.com/NightmareZero/nzgoutil/util"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 获取日志写入工具
// @param sync 同步写入?
// @param logPath 日志文件路径
// @param name 文件名前缀
func getFileWriter(sync bool, logPath string, name string) io.Writer {
	if sync {
		return getWriterSync(logPath, name)
	}
	return getWriterAsync(logPath, name)
}

// 获取同步写入工具
// @param logPath 日志文件路径
// @param name 文件名前缀
func getWriterSync(logPath string, name string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// 保存30天内的日志，每天分割一次日志
	hook, err := rotatelogs.New(
		uos.FixPathEndSlash(logPath)+name+"-%Y%m%d.log",
		rotatelogs.WithMaxAge(30*time.Hour*24),      // 保留30天
		rotatelogs.WithRotationTime(1*time.Hour*24), // 每天一次
		rotatelogs.WithRotationSize(3*1024*1024),    // 最长为3m
	)
	if err != nil {
		panic(err)
	}
	return hook
}

// 获取异步写入工具
// @param logPath 日志文件路径
// @param name 文件名前缀
func getWriterAsync(logPath string, name string) io.Writer {
	bws := zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(getWriterSync(logPath, name)),
		FlushInterval: 5 * time.Second,
	}
	return &bws
}

func getZapCores(config LogConfig, encoder zapcore.Encoder) (ret []zapcore.Core) {
	fixedPath := uos.FixPathEndSlash(util.If(len(config.Path) > 0, config.Path, "./log/"))

	// 初始化 log 文件输出
	logWriter := zapcore.AddSync(getFileWriter(config.Sync, fixedPath, "log"))
	ret = append(ret, zapcore.NewCore(
		encoder, logWriter, LogNormalLevel{config.Level, config.MergeErrorLog}))

	// 如果未开启日志合并
	if !config.MergeErrorLog {
		fixedErrPath := uos.FixPathEndSlash(util.If(len(config.ErrPath) > 0, config.ErrPath, "./log/"))
		errorLogWriter := zapcore.AddSync(getFileWriter(config.Sync, fixedErrPath, "err"))
		ret = append(ret, zapcore.NewCore(
			encoder, errorLogWriter, zap.ErrorLevel))
	}

	// 如果开启屏幕输出
	if config.Console {
		ret = append(ret, zapcore.NewCore(
			encoder, os.Stdout, LogNormalLevel{config.Level, false}))
		ret = append(ret, zapcore.NewCore(
			encoder, os.Stderr, zap.ErrorLevel))
	}

	return
}
