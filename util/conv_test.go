package util_test

import (
	"testing"

	"github.com/NightmareZero/nzgoutil/util"
)

type Tgc struct {
	Data string
}

func TestGobConv(t *testing.T) {
	var in = Tgc{
		Data: "data",
	}
	var out Tgc

	err := util.GobConv(in, &out)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if in.Data != out.Data {
		t.Errorf("data mismatch %v with %v", in.Data, out.Data)
	}
}
