package util

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

func MapConv(src map[string]any, dst any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			ToTimeHookFunc()),
		Result: dst,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(src); err != nil {
		return err
	}
	return err
}

func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(src reflect.Type, dst reflect.Type, data any) (any, error) {
		if dst != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch src.Kind() {
		case reflect.String:
			return FormatDateTime(data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func JsonConv(src any, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func GobConv(src any, dst any) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(src); err != nil {
		return err
	}

	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
