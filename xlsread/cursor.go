package xlsread

import (
	"fmt"
	"reflect"

	"github.com/xuri/excelize/v2"
)

type Cursor interface {
	Next() (hasNext bool)
	All(dst any) error
	Parse(dst any) error
}

type WriteCursor interface {
	Next()
	All(src any) error
	Format(src any) error
}

var _ Cursor = &RowReadCursor{}
var _ WriteCursor = &RowWriteCursor{}

// 读取光标
type RowReadCursor struct {
	row     int        // 行数
	data    [][]string // 数据
	parsers map[string]IParser
}

// 处理本行
// dst: 目标结构体
func (c *RowReadCursor) Parse(dst any) error {
	// 检查是否是指向结构体的指针
	if err := _isPtrTo(reflect.Struct, dst); err != nil {
		return fmt.Errorf("excelr: requires a pointer to struct as 'dst', %w ", err)
	}

	// 取出本行
	row := c.data[c.row-1]

	dstVal := reflect.ValueOf(dst).Elem()
	dstTyp := dstVal.Type()

	for i := 0; i < dstVal.NumField(); i++ {
		field := dstTyp.Field(i)
		fieldVal := dstVal.Field(i)

		// 跳过不可访问的字段 (如私有)
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			continue
		}

		// 获取标签信息
		tagInfo := getTagInfo(field.Tag.Get(TAG_NAME), c.parsers, nil)
		if 0 < tagInfo.col && tagInfo.col <= len(row) {
			s := row[tagInfo.col-1]

			var err error
			err = parseVal(s, fieldVal, tagInfo.parser)
			if err != nil {
				err = fmt.Errorf("error parse column %v '%v' to '%v' %w,", tagInfo.col, s, field.Name, err)
				return err
			}
		}

	}

	// t := reflect.TypeOf(dst)
	return nil
}

// funcNext implements Cursor
func (c *RowReadCursor) Next() (hasNext bool) {
	c.row++
	return (c.row - 1) < len(c.data)
}

// 将工作表中的数据根据struct中的tag插入结构体中 (目标为结构体切片)
func (c *RowReadCursor) All(dst any) error {

	// 检查是否是指向结构体的指针
	if err := _isPtrTo(reflect.Slice, dst); err != nil {
		return fmt.Errorf("excelr: requires a pointer to slice as 'dst', %w ", err)
	}

	dstVal := reflect.ValueOf(dst).Elem() // dst指向的类型 Kind = Slice

	sItemTyp := dstVal.Type().Elem()
	if sItemTyp.Kind() == reflect.Ptr {
		sItemTyp = sItemTyp.Elem()
	}

	// 遍历工作表行
	for c.Next() {
		// 创建新的结构体
		var pVal = reflect.New(sItemTyp).Interface()

		// 单行解析
		err := c.Parse(pVal)
		if err != nil {
			return fmt.Errorf("Document.UnmarshalRows: unmarshal error on row %v, %w", c.row, err)
		}

		// 插入目标数组
		dstVal.Set(reflect.Append(dstVal, reflect.ValueOf(pVal).Elem()))
	}
	return nil
}

type RowWriteCursor struct {
	h   *excelize.File // 文件句柄
	opt RowWriteOption // 配置
	row int            // 行数
	col int            // 最大列数(随着行数据变大而变大)
}

// All implements WriteCursor
func (c *RowWriteCursor) All(dst any) error {

	vDst := reflect.ValueOf(dst)
	if vDst.Type().Kind() == reflect.Ptr {
		vDst = vDst.Elem()
	}
	if vDst.Type().Kind() != reflect.Slice {
		return fmt.Errorf("excelr: requires a slice as 'src'")
	}

	for i := 0; i < vDst.Len(); i++ {
		// 创建新的结构体
		var pVal = vDst.Index(i).Interface()

		// 单行解析
		err := c.Format(pVal)
		if err != nil {
			return fmt.Errorf("Document.UnmarshalRows: unmarshal error on row %v, %w", c.row, err)
		}
		c.Next()
	}

	return nil
}

// Format implements WriteCursor
func (c *RowWriteCursor) Format(dst any) error {
	vDst := reflect.ValueOf(dst)
	if vDst.Type().Kind() == reflect.Ptr {
		vDst = vDst.Elem()
	}
	if vDst.Type().Kind() != reflect.Struct {
		return fmt.Errorf("excelr: requires a struct as 'src'")
	}

	dstVal := reflect.ValueOf(dst)
	dstTyp := dstVal.Type()

	for i := 0; i < dstVal.NumField(); i++ {
		field := dstTyp.Field(i)
		fieldVal := dstVal.Field(i)

		// 跳过不可访问的字段 (如私有)
		if !fieldVal.IsValid() {
			continue
		}

		// 获取标签信息
		tagInfo := getTagInfo(field.Tag.Get(TAG_NAME), nil, c.opt.Styles)

		if tagInfo.col > 0 {
			err := c.setVal(fieldVal.Interface(), i, tagInfo.style)
			if err != nil {
				err = fmt.Errorf("error set %v:%v  to column '%v' %w,", field.Name, fieldVal.Interface(), tagInfo.col, err)
				return err
			}
		}

	}

	return nil
}

func (c *RowWriteCursor) Next() {
	c.row++
}
