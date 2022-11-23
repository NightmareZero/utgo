package xlsread

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type TagInfo struct {
	row int
	col int
}

func getTagInfo(tag string) (res TagInfo) {
	if len(tag) == 0 {
		return
	}

	s := strings.Fields(tag)
	fieldMap := map[string]string{}
	for _, v := range s {
		nameAndVal := strings.Split(v, ":")
		if len(nameAndVal) > 1 {
			fieldMap[nameAndVal[0]] = nameAndVal[1]
		} else if len(nameAndVal) != 0 {
			fieldMap[nameAndVal[0]] = ""
		}
	}

	res.col, _ = strconv.Atoi(fieldMap["col"])
	res.row, _ = strconv.Atoi(fieldMap["row"])
	return
}

func parseVal(src string, dst any) error {
	k := reflect.TypeOf(dst).Kind()
	if reflect.Ptr != k {
		return errors.New("xlsr.parseVal: dst must be a pointer")
	}

	k = reflect.TypeOf(dst).Elem().Kind()
	v := reflect.ValueOf(dst).Elem()

	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(src, 10, 64)
		if err != nil {
			return fmt.Errorf("xlsr.parseVal: parse error, %w", err)
		}
		v.SetUint(i)
	case reflect.String:
		v.Set(reflect.ValueOf(src))
	}

	return nil
}
