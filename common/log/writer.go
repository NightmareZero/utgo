package log

import (
	"io"
	"time"

	"github.com/NightmareZero/m-go-starter/common/upath"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func getWriter(logPath string, name string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// 保存30天内的日志，每天分割一次日志
	hook, err := rotatelogs.New(
		upath.FixPathSlash(logPath)+name+"-%Y%m%d.log",
		rotatelogs.WithMaxAge(30*time.Hour*24),      // 保留30天
		rotatelogs.WithRotationTime(1*time.Hour*24), // 每天一次
		rotatelogs.WithRotationSize(3*1024*1024),    // 最长为1m
	)
	if err != nil {
		panic(err)
	}
	return hook
}
