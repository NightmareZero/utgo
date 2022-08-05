package common

import (
	"reflect"
	"strings"
	"unsafe"
)

func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

var splitHolder = []string{"戇", "⋚", "瘬", "🪃", "鈽", "〓", "艜", "嬔", "#", "$$", "の"}

// 使用分隔符与转义符对字符串进行分割
// $content 字符串内容
// $spliter 分隔符
// $escape 转义符
func SplitBy(content, spliter, escape string) (ret []string) {
	if len(spliter) == 0 {
		return []string{content}
	}
	if len(escape) == 0 {
		return strings.Split(content, spliter)
	}

	// 获取临时替换字符
	holder := ""
	for _, v := range splitHolder {
		if !strings.Contains(content, v) {
			if v != spliter && v != escape {
				holder = v
			}
		}
	}

	rcontent := strings.ReplaceAll(content, escape+spliter, holder)
	for _, v := range strings.Split(rcontent, spliter) {
		ret = append(ret, strings.ReplaceAll(v, holder, spliter))
	}

	return
}
