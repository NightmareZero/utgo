package common

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

// Try 保护方法运行
func Try(invoke func()) {
	// 延迟处理的函数
	defer Recover("panic", nil)
	invoke()
}

// TryCatch 保护方法运行，如果失败，则执行fallback方法
func TryCatch(invoke func(), fallback func(err error)) {
	defer Recover("panic", fallback)
	invoke()
}

// Recover 恢复panic
// 仅限于在 go routine 中，放置在第一行执行
// 严禁在主线程使用
func Recover(msg string, goAfterRecover func(err error)) {
	recoveredPanic := recover()
	if recoveredPanic == nil {
		return
	}
	doRecover(msg, recoveredPanic, goAfterRecover)
}

func doRecover(msg string, pan any, goAfterRecover func(err error)) {
	var err error
	buf := make([]byte, 4096)

	size := 0
	for {
		size = runtime.Stack(buf, false)
		// The size of the buffer may be not enough to hold the stacktrace,
		// so double the buffer size
		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break
	}
	switch ff := pan.(type) {
	case string:
		err = errors.New("panic: " + msg + " , " + ff + string(buf[:]))
	case error:
		err = errors.WithStack(ff)
	case uint, uint8, uint16, uint32, uint64,
		int, int8, int16, int32, int64:
		err = errors.Errorf("panic: error code %v, %v", ff, string(buf[:]))
	}
	if nil != goAfterRecover {
		Try(func() {
			goAfterRecover(err)
		})
	} else {
		fmt.Printf("%+v\n", err)
	}
}
