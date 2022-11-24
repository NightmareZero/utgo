package xlsread

import (
	"fmt"
	"reflect"
)

type Cursor interface {
	Next() (hasNext bool)
	All(dst any) error
	Parse(dst any) error
}

var _ Cursor = &RowReadCursor{}

// 读取光标
type RowReadCursor struct {
	row int // 行数
	// col   int        // 列数
	// sheet string     // 工作表
	data    [][]string //数据
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
		tagInfo := getTagInfo(c.parsers, field.Tag.Get(TAG_NAME))
		if 0 < tagInfo.col && tagInfo.col < len(row) {
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
