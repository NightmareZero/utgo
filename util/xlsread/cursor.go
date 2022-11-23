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
	Row   int        // 行数
	Col   int        // 列数
	Sheet string     // 工作表
	data  [][]string //数据
}

// Parse implements Cursor
func (c *RowReadCursor) Parse(dst any) error {
	// 检查是否是指向结构体的指针
	if !_isPtrTo(reflect.Struct, dst) {
		return fmt.Errorf("excelr: requires a pointer to struct as 'dst' ")
	}

	row := c.data[c.Row-1]

	dstTye := reflect.TypeOf(dst)
	for i := 0; i < dstTye.NumField(); i++ {
		field := dstTye.Field(i)
		tagInfo := getTagInfo(field.Tag.Get(TAG_NAME))
		if tagInfo.col < len(row) {
			s := row[tagInfo.col-1]
			pf := reflect.ValueOf(field).Addr()
			parseVal(s, pf)
		}

	}

	// t := reflect.TypeOf(dst)
	return nil
}

// funcNext implements Cursor
func (c *RowReadCursor) Next() (hasNext bool) {
	c.Row++
	return len(c.data) <= (c.Row - 1)
}

// 将工作表中的数据根据struct中的tag插入结构体中 (目标为结构体切片)
func (c *RowReadCursor) All(dst any) error {
	// 检查是否是指向结构体的指针
	if !(_isPtrTo(reflect.Slice, dst) || _isPtrTo(reflect.Array, dst)) {
		return fmt.Errorf("excelr: requires a pointer to slice as 'dst' ")
	}

	dstSliceVal := reflect.ValueOf(dst)

	// 遍历工作表行
	for c.Next() {
		// 创建新的结构体
		itemVal := reflect.ValueOf(dstSliceVal).Elem()
		var pDstRow any = reflect.New(itemVal.Type())

		// 单行解析
		err := c.Parse(pDstRow)
		if err != nil {
			return fmt.Errorf("Document.UnmarshalRows: unmarshal error on row %v, %w", c.Row, err)
		}

		// 插入目标数组
		reflect.Append(dstSliceVal, reflect.ValueOf(pDstRow))
	}
	return nil
}
