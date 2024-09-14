package log

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/NightmareZero/nzgoutil/vars"
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
		strings.TrimRight(logPath, vars.PATH_DELIMITER)+vars.PATH_DELIMITER+name+"-%Y%m%d.log",
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
	fixedPath := strings.TrimRight(If(len(config.Path) > 0, config.Path, "./log/"), vars.PATH_DELIMITER) + vars.PATH_DELIMITER

	if !config.NotToFile {
		// 初始化 log 文件输出
		logWriter := zapcore.AddSync(getFileWriter(config.Sync, fixedPath, "log"))
		ret = append(ret, zapcore.NewCore(
			encoder, logWriter, LogNormalLevel{config.level, config.MergeError}))

		// 如果未开启日志合并
		if !config.MergeError {
			fixedErrPath := strings.TrimRight(If(len(config.ErrPath) > 0, config.ErrPath, "./log/"), vars.PATH_DELIMITER) + vars.PATH_DELIMITER
			errorLogWriter := zapcore.AddSync(getFileWriter(config.Sync, fixedErrPath, "err"))
			ret = append(ret, zapcore.NewCore(
				encoder, errorLogWriter, zap.ErrorLevel))
		}
	}

	// 如果开启屏幕输出
	if config.Console {
		ret = append(ret, zapcore.NewCore(
			encoder, os.Stdout, LogNormalLevel{config.level, false}))
		ret = append(ret, zapcore.NewCore(
			encoder, os.Stderr, zap.ErrorLevel))
	}

	return
}

// 伪三元表达式
// PS: 请注意，由于go的机制, trueVal, falseVal 都是已定值(在传入if前就确定了), 可能会导致panic
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
