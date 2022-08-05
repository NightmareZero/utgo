package common_test

import (
	"reflect"
	"testing"

	"github.com/NightmareZero/nzgoutil/common"
)

func Test_SplitBy(t *testing.T) {
	tests := []struct {
		name    string
		content string
		spliter string
		escape  string
		assert  []string
	}{
		{
			name:    "t_dot",
			content: "1.2.3.4",
			spliter: ".",
			assert:  []string{"1", "2", "3", "4"},
		},
		{
			name:    "t_nospliter",
			content: "1.2.3.4",
			assert:  []string{"1.2.3.4"},
		},
		{
			name:    "t_with_escape",
			content: "1.2\\.3.4\\",
			spliter: ".",
			escape:  "\\",
			assert:  []string{"1", "2.3", "4\\"},
		}, {
			name:    "t_with_multi",
			content: "1and2and3\\and4",
			spliter: "and",
			escape:  "\\",
			assert:  []string{"1", "2", "3and4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := common.SplitBy(tt.content, tt.spliter, tt.escape)
			t.Logf("result %v", s)
			if !reflect.DeepEqual(s, tt.assert) {
				t.Errorf("result mismatch %v with %v", tt.assert, s)
			}
		})
	}
}
