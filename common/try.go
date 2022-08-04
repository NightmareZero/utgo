package common

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
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
	err := errors.New("panic: " + msg + " , type unknown")
	switch ff := recoveredPanic.(type) {
	case string:
		err = errors.New("panic: " + msg + " , " + ff)
	case error:
		err = errors.WithStack(ff)
	}
	if nil != goAfterRecover {
		Try(func() {
			goAfterRecover(err)
		})
	} else {
		fmt.Printf("%+v\n", zap.Error(err))
	}
}
