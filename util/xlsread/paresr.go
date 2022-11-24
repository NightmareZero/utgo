package xlsread

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrUnknownType = errors.New("unknown type field")
)

type TagInfo struct {
	row    int
	col    int
	parser IParser
}

type IParser func(string) any

func getTagInfo(parsers map[string]IParser, tag string) (res TagInfo) {
	if len(tag) == 0 {
		return
	}

	s := strings.Fields(tag)
	fieldMap := map[string]string{}
	for _, v := range s {
		nameAndVal := strings.Split(v, ":")
		if len(nameAndVal) > 1 {
			fieldMap[nameAndVal[0]] = nameAndVal[1]
		} else if len(nameAndVal) != 0 {
			fieldMap[nameAndVal[0]] = ""
		}
	}

	// 读取行列标签信息
	res.col, _ = strconv.Atoi(fieldMap[pColName])
	res.row, _ = strconv.Atoi(fieldMap[pRowName])

	// 读取处理器标签信息
	pParesrName := fieldMap[pParesr]
	res.parser = parsers[pParesrName]

	return
}

func parseVal(src string, dst reflect.Value, parser IParser) error {
	// k := reflect.TypeOf(dst).Kind()
	// if reflect.Ptr != k {
	// 	return errors.New("xlsr.parseVal: dst must be a pointer")
	// }

	k := dst.Kind()
	// v := reflect.ValueOf(dst).Elem()

	// 预处理有 Parser 的
	if parser != nil {
		parsed := parser(src)
		dst.Set(reflect.ValueOf(parsed))
		return nil
	}

	// 处理常规类型
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 处理数字类型
		i, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		dst.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 处理无符号数值类型
		i, err := strconv.ParseUint(src, 10, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		dst.SetUint(i)

	case reflect.String: // 处理字符串类型
		dst.Set(reflect.ValueOf(src))

	case reflect.Float32, reflect.Float64: // 处理浮点数类型
		f, err := strconv.ParseFloat(src, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		dst.SetFloat(f)

	case reflect.Bool: // 处理布尔类型
		lowerSrc := strings.ToLower(src)
		if lowerSrc == "true" || lowerSrc == "1" || lowerSrc == "yes" {
			dst.SetBool(true)
		}
	default:
		return ErrUnknownType

	}

	return nil
}
