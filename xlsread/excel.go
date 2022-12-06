package xlsread

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xuri/excelize/v2"
)

type Document struct {
	h *excelize.File // file handle
}

// Close implements io.Closer
func (d *Document) Close() (err error) {
	err = d.h.Close()
	d.h = nil
	return
}

func (d *Document) Save() (err error) {
	return d.h.Save()
}

func (d *Document) ReadSheetByRow(opt ...RowReadOption) (Cursor, error) {
	var opt1 RowReadOption = defaultRowReadOpt
	if len(opt) > 0 {
		opt1 = opt[0]
	}

	// 读取工作表数据
	sheetData, err := d.GetSheetData(&opt1.Option)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalRows.getSheetData, %w ", err)
	}

	// 拼装返回类型
	c := &RowReadCursor{}
	c.data = sheetData
	c.parsers = opt1.Parsers
	return c, nil
}

func (d *Document) WriteSheetByRow(src any, opt ...RowWriteOption) (WriteCursor, error) {
	var opt1 RowWriteOption = defaultRowWriteOpt
	if len(opt) > 0 {
		opt1 = opt[0]
	}

	// 拼装返回类型
	c := &RowWriteCursor{}
	c.h = d.h
	c.col = opt1.Col
	c.row = opt1.Row - 1 // 设置初始光标位置(data中)
	c.opt = opt1

	if c.row < 0 {
		c.row = 0
	}
	return c, nil
}

// 将工作表中的数据根据struct中的tag插入结构体中 (目标为单个结构体)
func (d *Document) ReadSheetByTable(dst any, opt ...*Option) error {
	// TODO 暂时不实现
	// 检查是否是指向结构体的指针
	if err := _isPtrTo(reflect.Slice, dst); err != nil {
		return fmt.Errorf("excelr: requires a pointer to slice as 'dst', %w ", err)
	}

	// t := reflect.TypeOf(dst)
	return nil
}

func (d *Document) GetSheetData(opt *Option) ([][]string, error) {
	return d.h.GetRows(opt.SheetName)
}

func (d *Document) WriteSheetData(opt *Option) (res [][]string, err error) {
	return
}

var (
	ErrNotPtr     = errors.New("dst needs to be a pointer")
	ErrInvlidKind = errors.New("invalid kind of dst")
)

// 是指向目标类型的指针
// isTyp: 预期的目标类型
// dst: 目标
func _isPtrTo(isTyp reflect.Kind, dst any) error {
	var dstTyp = reflect.TypeOf(dst)
	// 判断是否是指针类型
	if dstTyp.Kind() != reflect.Ptr {
		return ErrNotPtr
	}

	// 判断是否是目标类型
	if dstTyp.Elem().Kind() != isTyp {
		return fmt.Errorf("%w: %v != %v", ErrInvlidKind, dstTyp.Elem().Kind(), isTyp.String())
	}
	return nil
}
