package config

var (
	_iniReader = iniReader{}
)

type iniReader struct {
}

func (i *iniReader) Unmarshal(data []byte, v any) error {
	panic("unimplemented")
}
