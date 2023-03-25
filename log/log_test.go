package log_test

import (
	"testing"

	"github.com/NightmareZero/nzgoutil/log"
)

func Test_Info(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{
			name:  "test1",
			level: "debug",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.InitLog(tt.level)
			log.Debug("test")
			log.Info("test")
			log.Warn("test")
			log.Error("test")
		})
	}
}
