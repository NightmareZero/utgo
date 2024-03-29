package xlsread

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// 标签名称
const TAG_NAME = "xlsr"

// 标签项目
const (
	pRowName  = "row"   // 行标签
	pColName  = "col"   // 列标签
	pParesr   = "pp"    // 转换器标签
	pFormater = "style" // 格式标签
)

const (
	// 默认日期转换器
	defaultTimeFormat = "2006-01-02 15:04:05"
	defaultDateFormat = "2006-01-02"
)

// 打开一个excel文件
// path: 存储路径
// @return
// d: 操作句柄
func OpenFile(path string) (d *Document, e error) {
	doc := &Document{}
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("fail on open '%v', %w ", path, err)
	}
	doc.h = f
	return doc, nil
}

func OpenReader(reader io.Reader) (d *Document, e error) {
	doc := &Document{}
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("fail on open reader, %w ", err)
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

	return OpenFile(path)

}
