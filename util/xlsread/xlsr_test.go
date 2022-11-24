package xlsread

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type ProductItem struct {
	Name     string    `json:"name" xlsr:"col:1"`     // 品名
	LotNo    string    `json:"lotNo" xlsr:"col:2"`    // 批号
	TypeCode string    `json:"typeCode" xlsr:"col:3"` // 类型代码
	Quantity int16     `json:"quantity" xlsr:"col:4"` // 商品数量
	Price    float64   `json:"price" xlsr:"col:5"`    // 单价
	OutTime  time.Time `json:"outTime" xlsr:"col:6"`  // 过期时间
	InTime   time.Time `json:"inTime" xlsr:"col:7"`   // 进货时间

}

func Test_ReadSheet1(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		sheet string
		dst   []ProductItem
	}{
		{
			name:  "test-productItem",
			file:  "/home/user/nzgoutil/util/xlsread/tmp/Sheet1.xlsx",
			sheet: "Sheet1",
			dst:   []ProductItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := Open(tt.file)
			if err != nil {
				t.Error(fmt.Errorf("error opening file %v, %w", tt.file, err))
				return
			}

			rro := RowReadOption{}
			rro.SheetName = "Sheet1"

			cur, err := d.ReadSheetByRow(rro)
			if err != nil {
				t.Error(fmt.Errorf("error read excel file %v, %w", tt.file, err))
				return
			}

			cur.Next() // 跳过标题
			err = cur.All(&tt.dst)
			if err != nil {
				t.Error(fmt.Errorf("error parse file %v, %w", tt.file, err))
				return
			}

			t.Logf("haha: %+v", tt.dst)
		})
	}
}

func Test_A1(t *testing.T) {
	a := ProductItem{}
	av := reflect.ValueOf(a)
	for i := 0; i < av.NumField(); i++ {
		sf := av.Type().Field(i)
		f := av.Field(i)
		t.Logf("%v %v", sf.Name, f.Kind())
	}

}
