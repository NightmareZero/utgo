package xlsread

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

const TAG_NAME = "xlsr"

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
