package xlsread

import (
	"fmt"
	"reflect"

	"github.com/xuri/excelize/v2"
)

type Document struct {
	path string
	h    *excelize.File // file handle
}

func (d *Document) ReadSheetByRow(opt ...RowReadOption) (Cursor, error) {
	var opt1 RowReadOption
	if len(opt) > 0 {
		opt1 = opt[1]
	} else {
		opt1 = defaultRowReadOpt
	}

	// 读取工作表数据
	sheetData, err := d.GetSheetData(&opt1.Option)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalRows.getSheetData, %w ", err)
	}

	// 拼装返回类型
	c := &RowReadCursor{}
	c.data = sheetData
	return c, nil
}

// 将工作表中的数据根据struct中的tag插入结构体中 (目标为单个结构体)
func (d *Document) ReadSheetByTable(dst any, opt ...*Option) error {
	// TODO 暂时不实现
	// 检查是否是指向结构体的指针
	if !_isPtrTo(reflect.Struct, dst) {
		return fmt.Errorf("excelr: requires a pointer to struct as 'dst' ")
	}

	// t := reflect.TypeOf(dst)
	return nil
}

func (d *Document) GetSheetData(opt *Option) ([][]string, error) {
	return d.h.GetRows(opt.SheetName)
}

// 是指向目标类型的指针
// isTyp: 预期的目标类型
// dst: 目标
func _isPtrTo(isTyp reflect.Kind, dst any) bool {
	var dstTyp = reflect.TypeOf(dst)
	// 判断是否是指针类型
	if dstTyp.Kind() != reflect.Ptr {
		return false
	}

	// 判断是否是目标类型
	if dstTyp.Elem().Kind() == isTyp {
		return true
	}
	return false
}
