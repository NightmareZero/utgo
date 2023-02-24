package util_test

import (
	"reflect"
	"testing"

	common "github.com/NightmareZero/nzgoutil/util"
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

func Test_SplitHead(t *testing.T) {
	tests := []struct {
		name    string
		content string
		spliter string
		escape  string
		assert  [2]string
	}{
		{
			name:    "t_dot",
			content: "1.2.3.4",
			spliter: ".",
			assert:  [2]string{"1", "2.3.4"},
		},
		{
			name:    "t_nospliter",
			content: "1.2.3.4",
			assert:  [2]string{"1.2.3.4"},
		},
		{
			name:    "t_with_escape",
			content: "1.2\\.3.4\\",
			spliter: ".",
			escape:  "\\",
			assert:  [2]string{"1", "2.3.4\\"},
		}, {
			name:    "t_with_multi",
			content: "1\\and2and3and4",
			spliter: "and",
			escape:  "\\",
			assert:  [2]string{"1and2", "3and4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := common.SplitHead(tt.content, tt.spliter, tt.escape)
			t.Logf("result %v", s)
			if !reflect.DeepEqual(s, tt.assert) {
				t.Errorf("result mismatch %v with %v", tt.assert, s)
			}
		})
	}
}
