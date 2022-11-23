package xlsread

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// 标签名称
const TAG_NAME = "xlsr"

// 标签项目
const (
	pRowName = "row" // 行标签
	pColName = "col" // 列标签
	pParesr  = "pp"  // 转换器标签
)

const (
	// 默认日期转换器
	defaultTimeFormat = "2006-01-02 15:04:05"
)

// 打开一个excel文件
// path: 存储路径
// @return
// d: 操作句柄
func Open(path string) (d *Document, e error) {
	doc := &Document{
		path: path,
	}
	f, err := excelize.OpenFile(doc.path)
	if err != nil {
		return nil, fmt.Errorf("fail on open '%v', %w ", doc.path, err)
	}
	doc.h = f
	return doc, nil
}

// 新建一个excel文件
// path: 存储路径
// @return
// d: 操作句柄
func New(path string) (d *Document, e error) {
	f := excelize.NewFile()
	err := f.SaveAs(path)
	if err != nil {
		return nil, fmt.Errorf("fail on write '%v', %w", path, err)
	}

	return Open(path)

}
