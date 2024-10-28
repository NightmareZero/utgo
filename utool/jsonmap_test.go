package utool_test

import (
	"testing"

	"github.com/NightmareZero/nzgoutil/utool"
)

var _testData = map[string]any{
	"ms": map[string]any{
		"data": []map[string]any{
			{
				"id":   1,
				"name": "example",
			},
		},
		"name": "mysql",
	},
	"ms.name":   "mysql",
	"ms.0.name": "example",
	"list1": []map[string]any{
		{"name": "item3"},
		{"name": "item4", "ok": true},
	},
	"list": []any{
		map[string]any{"name": "item1"},
		map[string]any{"name": "item2"},
	},
}

func TestGetJsonMapValueByExpr(t *testing.T) {

	tests := []struct {
		expr     string
		expected any
		found    bool
	}{
		{"ms.data.0.name", "example", true},
		{"ms.data.id=1.name", "example", true},
		{"list.name=item1.name", "item1", true},
		{"ms.name", "mysql", true},
		{"list.0.name", "item1", true},
		{"list.1.name", "item2", true},
		{"list.2.name", nil, false},
		{"list1.name=item3.name", "item3", true},
		{"list1.ok=true.name", "item4", true},
		{"ms.1.name", nil, false},
		{"ms\\.name", "mysql", true},
		{"ms\\.0\\.name", "example", true},
	}

	for _, test := range tests {
		result, found := utool.JsonMapExpr.GetJsonMapValueByExpr(test.expr, _testData)
		if found != test.found || result != test.expected {
			t.Errorf("GetJsonMapValueByExpr(%q) = (%v, %v), expected (%v, %v)", test.expr, result, found, test.expected, test.found)
		}
	}
}

func TestSplitKey(t *testing.T) {
	tests := []struct {
		expr     string
		expected []string
	}{
		{"ms.0.name", []string{"ms", "0", "name"}},
		{"ms\\.name", []string{"ms.name"}},
		{"ms\\.0\\.name", []string{"ms.0.name"}},
		{"list.0.name", []string{"list", "0", "name"}},
	}

	for _, test := range tests {
		result := utool.JsonMapExpr.SplitKey(test.expr)
		if len(result) != len(test.expected) {
			t.Errorf("splitKey(%q) = %v, expected %v", test.expr, result, test.expected)
			continue
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("splitKey(%q) = %v, expected %v", test.expr, result, test.expected)
				break
			}
		}
	}
}

func TestSplitEqual(t *testing.T) {
	tests := []struct {
		expr        string
		expectedKey string
		expectedVal string
	}{
		{"ms.name=mysql.name", "ms.name", "mysql.name"},
		{"ms\\.name=mysql\\.name", "ms.name", "mysql.name"},
		{"key=value", "key", "value"},
		{"key\\=value=val", "key=value", "val"},
		{"key=value=val", "key", "value=val"},
		{"key\\=value", "key=value", ""},
		{"ms", "ms", ""},
	}

	for _, test := range tests {
		key, val := utool.JsonMapExpr.SplitEqual(test.expr)
		if key != test.expectedKey || val != test.expectedVal {
			t.Errorf("splitEqual(%q) = (%q, %q), expected (%q, %q)", test.expr, key, val, test.expectedKey, test.expectedVal)
		}
	}
}
