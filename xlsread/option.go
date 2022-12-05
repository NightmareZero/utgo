package xlsread

type Option struct {
	SheetName string
}

type RowReadOption struct {
	Parsers map[string]IParser
	Option
}

type RowWriteOption struct {
	Option
	Col       int // 最大列数(过多会导致内存浪费，过少会导致数据被截断)
	Row       int // 启示行数
	Formaters map[string]IFormater
}

// 默认单表读取选项
var (
	defaultSingleSheetOpt = Option{
		SheetName: "Sheet1",
	}
	// 默认单行读取选项
	defaultRowReadOpt = RowReadOption{
		Option: defaultSingleSheetOpt,
	}
	defaultRowWriteOpt = RowWriteOption{
		Col:    24,
		Option: defaultSingleSheetOpt,
	}
)
