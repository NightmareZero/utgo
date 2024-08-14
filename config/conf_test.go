package config_test

import (
	"reflect"
	"testing"

	"github.com/NightmareZero/nzgoutil/config"
	"github.com/NightmareZero/nzgoutil/utilp"
)

func Test_SplitBy(t *testing.T) {
	tests := []struct {
		name    string
		content string
		spliter string
		escape  string
	}{
		{
			name:    "t_dot",
			content: "1.2.3.4",
			spliter: ".",
		},
		{
			name:    "t_nospliter",
			content: "1.2.3.4",
		},
		{
			name:    "t_with_escape",
			content: "1.2\\.3.4\\",
			spliter: ".",
			escape:  "\\",
		}, {
			name:    "t_with_multi",
			content: "1and2and3\\and4",
			spliter: "and",
			escape:  "\\",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := utilp.SplitBy(tt.content, tt.spliter, tt.escape)
			t.Logf("result %v", s)
		})
	}
}

func Test_Config(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		confType config.ConfigType
		assert   map[string]any
	}{
		{
			name:     "yaml",
			path:     "testYaml.yaml",
			confType: config.Yaml,
			assert: map[string]any{
				"str":       "str",
				"number":    3,
				"obj.array": []any{"1", "2"},
			},
		},
		{
			name:     "json",
			path:     "testJson.json",
			confType: config.Json,
			assert: map[string]any{
				"str":       "str",
				"number":    float64(3),
				"obj.array": []any{"1", "2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			myconf, err := config.FromFile(tt.path, tt.confType)
			if err != nil {
				t.Error(err)
				return
			}
			for i, v := range tt.assert {
				ret, ok := myconf.Val(i)
				if !ok {
					t.Errorf("%v not found", i)
					continue
				}
				if !reflect.DeepEqual(ret, v) {
					t.Errorf("%v %v not equal assert %v", i, ret, v)
				}
			}
		})
	}
}
