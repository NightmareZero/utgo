package xlsread

type Option struct {
	SheetName string
	Parsers   map[string]IParser
}

type RowReadOption struct {
	Option
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
)
