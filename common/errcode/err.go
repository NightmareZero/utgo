package errcode

import "fmt"

var (
	Success = ErrorCode{0, "", nil}
	Failure = ErrorCode{-1, "未知错误", nil}
)

type ErrorCode struct {
	Code  int
	Msg   string
	cause error
}

func (e ErrorCode) Error() string {
	return fmt.Sprintf("%v: %v, %+v", e.Code, e.Msg, e.cause)
}

func (e *ErrorCode) MsgOf(s string) string {
	if len(e.Msg) > 0 && len(s) > 0 {
		return s + ":  " + e.Error()
	} else if len(e.Msg) > 0 {
		return e.Error()
	} else {
		return s
	}
}

func (e ErrorCode) Warp(err error) ErrorCode {
	e.cause = err
	return e
}

func (e ErrorCode) Warpf(format string, val ...any) ErrorCode {
	e.cause = fmt.Errorf(format, val...)
	return e
}

var _ error = &ErrorCode{}
