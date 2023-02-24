package util

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

	// 将转义替换为临时字符
	rcontent := strings.ReplaceAll(content, escape+spliter, holder)

	// 执行拆分
	for _, v := range strings.Split(rcontent, spliter) {
		ret = append(ret, strings.ReplaceAll(v, holder, spliter))
	}

	return
}

// 使用分隔符与转义符对字符串进行分割(只拆分出第一个)
// $content 字符串内容
// $spliter 分隔符
// $escape 转义符
func SplitHead(content, spliter, escape string) (ret [2]string) {
	if len(spliter) == 0 {
		return [2]string{content}
	}
	if len(escape) == 0 {
		s := strings.SplitN(content, spliter, 2)
		if len(s) > 1 {
			return [2]string{s[0], s[1]}
		}
		return [2]string{content}
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

	// 将转义替换为临时字符
	rcontent := strings.ReplaceAll(content, escape+spliter, holder)

	// 执行拆分
	s := strings.SplitN(rcontent, spliter, 2)
	if len(s) > 1 {
		return [2]string{strings.ReplaceAll(s[0], holder, spliter), strings.ReplaceAll(s[1], holder, spliter)}
	}
	return [2]string{strings.ReplaceAll(s[0], holder, spliter)}
}
