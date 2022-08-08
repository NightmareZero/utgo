package log

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/NightmareZero/nzgoutil/common"
	"github.com/NightmareZero/nzgoutil/uos"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
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
	return newAsyncWriter(logPath, name)
}

// 异步日志写入工具
type asyncWriter struct {
	syncLogger io.Writer
	ch         chan []byte
	ctx        context.Context
	cf         context.CancelFunc
	err        error
}

// 异步写入数据
// @Override
func (a *asyncWriter) Write(b []byte) (int, error) {
	builder := strings.Builder{}
	go common.Try(func() {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-t.C:
				if builder.Len() > 0 {
					a.ch <- []byte(builder.String())
					builder.Reset()
				}
			}
		}
	})

	select {
	case <-a.ctx.Done():
		return 0, errors.New("writer is closed")
	default:
		return builder.Write(b)
	}
}

// 关闭通道
// @Override
func (a *asyncWriter) Close() error {
	a.cf()
	return nil
}

// 新建异步写入通道
// newAsyncWriter
func newAsyncWriter(logPath string, name string) io.WriteCloser {
	var w = &asyncWriter{}
	w.ch = make(chan []byte, 1024)
	w.syncLogger = getWriterSync(logPath, name)
	w.ctx, w.cf = context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				return
			case b1 := <-w.ch:
				_, err := w.syncLogger.Write(b1)
				if err != nil {
					w.err = err
				}
			}
		}
	}()

	return w
}
