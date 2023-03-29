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
			err := log.InitLog(tt.level)
			if err != nil {
				t.Error(err)
			}
			log.Debugf("test")
			log.Current.Debugf("test")
			log.Infof("test")
			log.Current.Infof("test")
			log.Warnf("test")
			log.Current.Warnf("test")
			log.Errorf("test")
			log.Current.Errorf("test")
		})
	}
}
