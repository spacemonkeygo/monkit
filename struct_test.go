package monkit

import (
	"reflect"
	"testing"
)

func TestStatSourceFromStruct(t *testing.T) {
	type SubStruct struct {
		SubBool  bool
		SubFloat float64
		SubInt   int64
		Ignored  float64 `monkit:"ignore"`
	}

	result := Collect(StatSourceFromStruct(NewSeriesKey("struct"),
		struct {
			SomeBool  bool
			SomeFloat float64
			SomeInt   int64
			Ignored   float64 `monkit:"ignore"`
			Sub       SubStruct
			Skip      struct {
				Nope int64
			} `monkit:"whatever,ignore"`
		}{
			SomeInt:  5,
			SomeBool: true,
			Sub: SubStruct{
				SubFloat: 3.2,
			},
		},
	))

	if !reflect.DeepEqual(result, map[string]float64{
		"struct SomeBool":     1,
		"struct SomeFloat":    0,
		"struct SomeInt":      5,
		"struct Sub.SubBool":  0,
		"struct Sub.SubFloat": 3.2,
		"struct Sub.SubInt":   0,
	}) {
		t.Fatal("unexpected result", result)
	}
}
