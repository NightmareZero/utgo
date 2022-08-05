package config

var (
	_propertiesReader = propertiesReader{}
)

type propertiesReader struct {
}

func (r *propertiesReader) Unmarshal(data []byte, v any) error {
	panic("unimplemented")
}
