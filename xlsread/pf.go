package xlsread

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var (
	ErrUnknownType = errors.New("unknown type field")
)

type TagInfo struct {
	name   string
	row    int
	col    int
	parser IParser
	style  *CellStyle
}

type IParser func(string) any

// 默认日期格式处理
func DefaultDateParser(src string) any {
	excelDate := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	var days, _ = strconv.Atoi(src)
	return excelDate.Add(time.Second * time.Duration(days*86400))
}

func DefaultStrDataParser(src string) any {
	src = strings.ReplaceAll(src, "/", "-")
	src = strings.ReplaceAll(src, "\\", "-")
	t, err := time.Parse(defaultDateFormat, src)
	if err == nil {
		return t
	}

	return time.Now()
}

// 格式转换器
type CellStyle struct {
	Style *excelize.Style
	id    int
}

var (
	DefaultDateStyleStr = "yyyy-mm-dd"
	DefaultDateStyle    = CellStyle{
		Style: &excelize.Style{CustomNumFmt: &DefaultDateStyleStr},
	}
)

func getTagInfo(name string, tag string, parsers map[string]IParser, styles map[string]CellStyle) (res TagInfo) {
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
	res.name = name
	res.col, _ = strconv.Atoi(fieldMap[pColName])
	res.row, _ = strconv.Atoi(fieldMap[pRowName])

	// 读取处理器标签信息
	if parsers != nil {
		pParesrName := fieldMap[pParesr]
		res.parser = parsers[pParesrName]
	}

	if styles != nil {
		pFormaterName := fieldMap[pFormater]
		cs := styles[pFormaterName]
		res.style = &cs
	}

	return
}

func parseVal(src string, dst reflect.Value, parser IParser) error {
	// 预处理有 Parser 的
	if parser != nil {
		parsed := parser(src)
		dst.Set(reflect.ValueOf(parsed))
		return nil
	}

	// 处理常规类型
	switch dst.Interface().(type) {
	case int, int8, int16, int32, int64:
		// 处理数字类型
		i, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		dst.SetInt(i)

	case uint, uint8, uint16, uint32, uint64:
		// 处理无符号数值类型
		i, err := strconv.ParseUint(src, 10, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		dst.SetUint(i)

	case string: // 处理字符串类型
		dst.Set(reflect.ValueOf(src))

	case float32, float64: // 处理浮点数类型
		f, err := strconv.ParseFloat(src, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		dst.SetFloat(f)

	case bool: // 处理布尔类型
		lowerSrc := strings.ToLower(src)
		if lowerSrc == "true" || lowerSrc == "1" || lowerSrc == "yes" || lowerSrc == "是" {
			dst.SetBool(true)
		}

	case time.Time: // 处理日期类型
		var t any
		if strings.Contains(src, "-") || strings.Contains(src, "\\") || strings.Contains(src, "/") {
			t = DefaultStrDataParser(src)
		} else {
			t = DefaultDateParser(src)
		}
		dst.Set(reflect.ValueOf(t))

	default:
		return ErrUnknownType

	}

	return nil
}

func (c *RowWriteCursor) setVal(src any, col int, style *CellStyle) (err error) {

	axis, _ := excelize.CoordinatesToCellName(col, c.row+1)
	// 预处理有 Parser 的
	// if formater != nil {
	// 	return formater(src)
	// }

	// 处理常规类型
	switch tsrc := src.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool:
		// 处理数字类型
		return c.h.SetCellValue(c.opt.SheetName, axis, tsrc)
	case time.Time: // 处理日期类型
		if style == nil {
			style = &DefaultDateStyle
		}
		if style.id == 0 {
			style.id, err = c.h.NewStyle(style.Style)
			if err != nil {
				return err
			}
		}

		c.h.SetCellValue(c.opt.SheetName, axis, tsrc)
		c.h.SetCellStyle(c.opt.SheetName, axis, axis, style.id)

		return nil
	default:
		return ErrUnknownType

	}
}
