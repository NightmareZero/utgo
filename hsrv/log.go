package hsrv

import (
	"fmt"
	"io"
	"os"
	"time"
)

const defaultLogTimeFormatter = "2006-01-02 15:04:05"
const (
	LvlDebug = 0
	LvlInfo  = 1
	LvlWarn  = 2
	LvlError = 3
)

var LvlName = [4]string{"DEBUG", "INFO", "WARN", "ERROR"}

type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warn(v ...any)
	Error(v ...any)
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
}

var defaultLogger Logger = &_defaultLogger{
	Level: 0,
	Out:   os.Stdout,
}

type _defaultLogger struct {
	Level int
	Out   io.Writer
}

// Debug implements Logger
func (c *_defaultLogger) Debug(v ...any) {
	c.print(0, v...)
}

// Debugf implements Logger
func (c *_defaultLogger) Debugf(format string, v ...any) {
	c.print(0, fmt.Sprintf(format, v...))
}

// Error implements Logger
func (c *_defaultLogger) Error(v ...any) {
	c.print(3, v...)
}

// Errorf implements Logger
func (c *_defaultLogger) Errorf(format string, v ...any) {
	c.print(3, fmt.Sprintf(format, v...))
}

// Info implements Logger
func (c *_defaultLogger) Info(v ...any) {
	c.print(1, v...)
}

// Infof implements Logger
func (c *_defaultLogger) Infof(format string, v ...any) {
	c.print(1, fmt.Sprintf(format, v...))
}

// Warn implements Logger
func (c *_defaultLogger) Warn(v ...any) {
	c.print(2, v...)
}

// Warnf implements Logger
func (c *_defaultLogger) Warnf(format string, v ...any) {
	c.print(2, fmt.Sprintf(format, v...))
}

func (c *_defaultLogger) print(lvl int, msg ...any) {
	if c.Level <= lvl {
		timeStr := time.Now().Format(defaultLogTimeFormatter)
		_, _ = fmt.Fprintln(c.Out, fmt.Sprintf("%v [%v] %+v\n", timeStr, LvlName[lvl], msg))
	}
}
