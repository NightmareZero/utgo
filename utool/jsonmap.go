package utool

/*
json项表达式获取
最外层必定是 map[string]any
支持基础结构 map[string]any 和 []any
子项字符为 "." 转义字符为 "\."
判断字符为 "=" 转义字符为 "\="

例如:
ms.0.name 直接获取项
ms.name=mysql.name 根据条件获取项
*/

import (
	"strconv"
	"strings"

	"github.com/NightmareZero/nzgoutil/utilp"
)

var JsonMapExpr = JsonMapExprGetter{}

type JsonMapExprGetter struct{}

// GetJsonMapValueByExpr 从 map[string]any 中根据表达式获取项
func (g JsonMapExprGetter) GetJsonMapValueByExpr(expr string, data map[string]any) (any, bool) {
	parts := g.SplitKey(expr)
	var current any = data

	for _, part := range parts {
		unescapedPart := strings.ReplaceAll(part, "\\.", ".")
		if g.isCondition(unescapedPart) {
			switch v := current.(type) {
			case map[string]any:
				current, _ = g.getValueByCondition(unescapedPart, v)
			case []map[string]any:
				for _, item := range v {
					if val, ok := g.getValueByCondition(unescapedPart, item); ok {
						current = val
						break
					}
				}
			case []any:
				for _, item := range v {
					if val, ok := item.(map[string]any); ok {
						if val, ok := g.getValueByCondition(unescapedPart, val); ok {
							current = val
							break
						}
					}
				}
			default:
				return nil, false
			}
		} else {
			switch v := current.(type) {
			case map[string]any:

				if val, ok := v[unescapedPart]; ok {
					current = val
				} else {
					return nil, false
				}

			case []map[string]any:
				index, err := strconv.Atoi(part)
				if err != nil || index < 0 || index >= len(v) {
					return nil, false
				}
				current = v[index]
			case []any:
				index, err := strconv.Atoi(part)
				if err != nil || index < 0 || index >= len(v) {
					return nil, false
				}
				current = v[index]
			default:
				return nil, false
			}
		}
	}
	return current, true
}

// 判断是不是 equal 表达式
func (g JsonMapExprGetter) isCondition(expr string) bool {
	var _, val = g.SplitEqual(expr)
	return val != ""
}

// getValueByCondition 根据条件从 map[string]any 中获取值
func (g JsonMapExprGetter) getValueByCondition(expr string, data map[string]any) (any, bool) {
	key, val := g.SplitEqual(expr)
	if itemVal, ok := data[key]; ok {
		if utilp.ToStr(itemVal) == val {
			return data, true
		}
	}
	return nil, false
}

// SplitKey 将表达式按 '.' 分割，处理转义字符 '\.'
func (JsonMapExprGetter) SplitKey(expr string) []string {
	var parts []string
	var currentPart strings.Builder

	escaped := false
	for _, char := range expr {
		if escaped {
			currentPart.WriteRune(char)
			escaped = false
		} else {
			if char == '\\' {
				escaped = true
			} else if char == '.' {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			} else {
				currentPart.WriteRune(char)
			}
		}
	}
	parts = append(parts, currentPart.String())
	return parts
}

// SplitEqual 将表达式按 '=' 分割，处理转义字符 '\='
// 如果没有一个非转义的 '=' 则将整个表达式作为 key, val 为空
func (JsonMapExprGetter) SplitEqual(expr string) (key, val string) {
	var currentPart strings.Builder

	escaped := false    // 有转义字符
	foundEqual := false // 是否找到等号

	for _, char := range expr {
		if escaped {
			currentPart.WriteRune(char)
			escaped = false
		} else {
			if char == '\\' {
				escaped = true
			} else if char == '=' {
				if !foundEqual {
					key = currentPart.String()
					currentPart.Reset()
					foundEqual = true
				} else {
					currentPart.WriteRune(char)
				}
			} else {
				currentPart.WriteRune(char)
			}
		}
	}
	if !foundEqual {
		key = currentPart.String()
	} else {
		val = currentPart.String()
	}
	return
}
