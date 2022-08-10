package log

import (
	"io"
	"time"

	"github.com/NightmareZero/nzgoutil/uos"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
)

// 获取日志写入工具
// @param sync 同步写入?
// @param logPath 日志文件路径
// @param name 文件名前缀
func getWriter(sync bool, logPath string, name string) io.Writer {
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
