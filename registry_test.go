package jsonapi_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	. "github.com/mark-hartmann/jsonapi"
	"github.com/stretchr/testify/assert"
)

func TestGetZeroValue(t *testing.T) {
	testData := map[string][]struct {
		val       interface{}
		typ       int
		arr, null bool
		err       bool
	}{
		"default": {
			{val: "", typ: AttrTypeString, arr: false, null: false},
			{val: int(0), typ: AttrTypeInt, arr: false, null: false},
			{val: int8(0), typ: AttrTypeInt8, arr: false, null: false},
			{val: int16(0), typ: AttrTypeInt16, arr: false, null: false},
			{val: int32(0), typ: AttrTypeInt32, arr: false, null: false},
			{val: int64(0), typ: AttrTypeInt64, arr: false, null: false},
			{val: uint(0), typ: AttrTypeUint, arr: false, null: false},
			{val: uint8(0), typ: AttrTypeUint8, arr: false, null: false},
			{val: uint16(0), typ: AttrTypeUint16, arr: false, null: false},
			{val: uint32(0), typ: AttrTypeUint32, arr: false, null: false},
			{val: uint64(0), typ: AttrTypeUint64, arr: false, null: false},
			{val: float32(0), typ: AttrTypeFloat32, arr: false, null: false},
			{val: float64(0), typ: AttrTypeFloat64, arr: false, null: false},
			{val: false, typ: AttrTypeBool, arr: false, null: false},
			{val: time.Time{}, typ: AttrTypeTime, arr: false, null: false},
		},
		"array": {
			{val: []string{}, typ: AttrTypeString, arr: true, null: false},
			{val: []int{}, typ: AttrTypeInt, arr: true, null: false},
			{val: []int8{}, typ: AttrTypeInt8, arr: true, null: false},
			{val: []int16{}, typ: AttrTypeInt16, arr: true, null: false},
			{val: []int32{}, typ: AttrTypeInt32, arr: true, null: false},
			{val: []int64{}, typ: AttrTypeInt64, arr: true, null: false},
			{val: []uint{}, typ: AttrTypeUint, arr: true, null: false},
			{val: []uint8{}, typ: AttrTypeUint8, arr: true, null: false},
			{val: []uint16{}, typ: AttrTypeUint16, arr: true, null: false},
			{val: []uint32{}, typ: AttrTypeUint32, arr: true, null: false},
			{val: []uint64{}, typ: AttrTypeUint64, arr: true, null: false},
			{val: []float32{}, typ: AttrTypeFloat32, arr: true, null: false},
			{val: []float64{}, typ: AttrTypeFloat64, arr: true, null: false},
			{val: []bool{}, typ: AttrTypeBool, arr: true, null: false},
			{val: []time.Time{}, typ: AttrTypeTime, arr: true, null: false},
		},
		"nullable": {
			{val: (*string)(nil), typ: AttrTypeString, arr: false, null: true},
			{val: (*int)(nil), typ: AttrTypeInt, arr: false, null: true},
			{val: (*int8)(nil), typ: AttrTypeInt8, arr: false, null: true},
			{val: (*int16)(nil), typ: AttrTypeInt16, arr: false, null: true},
			{val: (*int32)(nil), typ: AttrTypeInt32, arr: false, null: true},
			{val: (*int64)(nil), typ: AttrTypeInt64, arr: false, null: true},
			{val: (*uint)(nil), typ: AttrTypeUint, arr: false, null: true},
			{val: (*uint8)(nil), typ: AttrTypeUint8, arr: false, null: true},
			{val: (*uint16)(nil), typ: AttrTypeUint16, arr: false, null: true},
			{val: (*uint32)(nil), typ: AttrTypeUint32, arr: false, null: true},
			{val: (*uint64)(nil), typ: AttrTypeUint64, arr: false, null: true},
			{val: (*float32)(nil), typ: AttrTypeFloat32, arr: false, null: true},
			{val: (*float64)(nil), typ: AttrTypeFloat64, arr: false, null: true},
			{val: (*bool)(nil), typ: AttrTypeBool, arr: false, null: true},
			{val: (*time.Time)(nil), typ: AttrTypeTime, arr: false, null: true},
		},
		"nullable array": {
			{val: (*[]string)(nil), typ: AttrTypeString, arr: true, null: true},
			{val: (*[]int)(nil), typ: AttrTypeInt, arr: true, null: true},
			{val: (*[]int8)(nil), typ: AttrTypeInt8, arr: true, null: true},
			{val: (*[]int16)(nil), typ: AttrTypeInt16, arr: true, null: true},
			{val: (*[]int32)(nil), typ: AttrTypeInt32, arr: true, null: true},
			{val: (*[]int64)(nil), typ: AttrTypeInt64, arr: true, null: true},
			{val: (*[]uint)(nil), typ: AttrTypeUint, arr: true, null: true},
			{val: (*[]uint8)(nil), typ: AttrTypeUint8, arr: true, null: true},
			{val: (*[]uint16)(nil), typ: AttrTypeUint16, arr: true, null: true},
			{val: (*[]uint32)(nil), typ: AttrTypeUint32, arr: true, null: true},
			{val: (*[]uint64)(nil), typ: AttrTypeUint64, arr: true, null: true},
			{val: (*[]float32)(nil), typ: AttrTypeFloat32, arr: true, null: true},
			{val: (*[]float64)(nil), typ: AttrTypeFloat64, arr: true, null: true},
			{val: (*[]bool)(nil), typ: AttrTypeBool, arr: true, null: true},
			{val: (*[]time.Time)(nil), typ: AttrTypeTime, arr: true, null: true},
		},
		"bytes": {
			{val: []uint8{}, typ: AttrTypeBytes, arr: false, null: false},
			{val: []uint8{}, typ: AttrTypeBytes, arr: true, null: false},
			{val: (*[]uint8)(nil), typ: AttrTypeBytes, arr: false, null: true},
			{val: (*[]uint8)(nil), typ: AttrTypeBytes, arr: true, null: true},
		},
		"invalid types": {
			{val: nil, typ: AttrTypeInvalid, err: true},
			{val: nil, typ: 99999, err: true},
		},
	}

	for name, tests := range testData {
		t.Run(name, func(t *testing.T) {
			for _, test := range tests {
				zv, err := GetZeroValue(test.typ, test.arr, test.null)

				assert.Equal(t, test.val, zv)
				if test.err {
					assert.Error(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestUnmarshalToType(t *testing.T) {
	var (
		vstr     = "str"
		vint     = int(1)
		vint8    = int8(8)
		vint16   = int16(16)
		vint32   = int32(32)
		vint64   = int64(64)
		vuint    = uint(1)
		vuint8   = uint8(8)
		vuint16  = uint16(16)
		vuint32  = uint32(32)
		vuint64  = uint64(64)
		vfloat32 = float32(math.MaxFloat32)
		vfloat64 = math.MaxFloat64
		vbool    = true
		vtime    = time.Time{}

		vstrarr     = []string{"str"}
		vintarr     = []int{1}
		vint8arr    = []int8{8}
		vint16arr   = []int16{16}
		vint32arr   = []int32{32}
		vint64arr   = []int64{64}
		vuintarr    = []uint{1}
		vuint8arr   = []uint8{8}
		vuint16arr  = []uint16{16}
		vuint32arr  = []uint32{32}
		vuint64arr  = []uint64{64}
		vfloat32arr = []float32{math.MaxFloat32}
		vfloat64arr = []float64{math.MaxFloat64}
		vboolarr    = []bool{true}
		vtimearr    = []time.Time{{}}
	)

	testData := map[string][]struct {
		val  interface{}
		attr Attr
		err  bool
	}{
		"default": {
			{val: vstr, attr: Attr{Type: AttrTypeString, Array: false, Nullable: false}},
			{val: vint, attr: Attr{Type: AttrTypeInt, Array: false, Nullable: false}},
			{val: vint8, attr: Attr{Type: AttrTypeInt8, Array: false, Nullable: false}},
			{val: vint16, attr: Attr{Type: AttrTypeInt16, Array: false, Nullable: false}},
			{val: vint32, attr: Attr{Type: AttrTypeInt32, Array: false, Nullable: false}},
			{val: vint64, attr: Attr{Type: AttrTypeInt64, Array: false, Nullable: false}},
			{val: vuint, attr: Attr{Type: AttrTypeUint, Array: false, Nullable: false}},
			{val: vuint8, attr: Attr{Type: AttrTypeUint8, Array: false, Nullable: false}},
			{val: vuint16, attr: Attr{Type: AttrTypeUint16, Array: false, Nullable: false}},
			{val: vuint32, attr: Attr{Type: AttrTypeUint32, Array: false, Nullable: false}},
			{val: vuint64, attr: Attr{Type: AttrTypeUint64, Array: false, Nullable: false}},
			{val: vfloat32, attr: Attr{Type: AttrTypeFloat32, Array: false, Nullable: false}},
			{val: vfloat64, attr: Attr{Type: AttrTypeFloat64, Array: false, Nullable: false}},
			{val: vbool, attr: Attr{Type: AttrTypeBool, Array: false, Nullable: false}},
			{val: vtime, attr: Attr{Type: AttrTypeTime, Array: false, Nullable: false}},
		},
		"array": {
			{val: vstrarr, attr: Attr{Type: AttrTypeString, Array: true, Nullable: false}},
			{val: vintarr, attr: Attr{Type: AttrTypeInt, Array: true, Nullable: false}},
			{val: vint8arr, attr: Attr{Type: AttrTypeInt8, Array: true, Nullable: false}},
			{val: vint16arr, attr: Attr{Type: AttrTypeInt16, Array: true, Nullable: false}},
			{val: vint32arr, attr: Attr{Type: AttrTypeInt32, Array: true, Nullable: false}},
			{val: vint64arr, attr: Attr{Type: AttrTypeInt64, Array: true, Nullable: false}},
			{val: vuintarr, attr: Attr{Type: AttrTypeUint, Array: true, Nullable: false}},
			{val: vuint8arr, attr: Attr{Type: AttrTypeUint8, Array: true, Nullable: false}},
			{val: vuint16arr, attr: Attr{Type: AttrTypeUint16, Array: true, Nullable: false}},
			{val: vuint32arr, attr: Attr{Type: AttrTypeUint32, Array: true, Nullable: false}},
			{val: vuint64arr, attr: Attr{Type: AttrTypeUint64, Array: true, Nullable: false}},
			{val: vfloat32arr, attr: Attr{Type: AttrTypeFloat32, Array: true, Nullable: false}},
			{val: vfloat64arr, attr: Attr{Type: AttrTypeFloat64, Array: true, Nullable: false}},
			{val: vboolarr, attr: Attr{Type: AttrTypeBool, Array: true, Nullable: false}},
			{val: vtimearr, attr: Attr{Type: AttrTypeTime, Array: true, Nullable: false}},
		},
		"nullable": {
			{val: &vstr, attr: Attr{Type: AttrTypeString, Array: false, Nullable: true}},
			{val: &vint, attr: Attr{Type: AttrTypeInt, Array: false, Nullable: true}},
			{val: &vint8, attr: Attr{Type: AttrTypeInt8, Array: false, Nullable: true}},
			{val: &vint16, attr: Attr{Type: AttrTypeInt16, Array: false, Nullable: true}},
			{val: &vint32, attr: Attr{Type: AttrTypeInt32, Array: false, Nullable: true}},
			{val: &vint64, attr: Attr{Type: AttrTypeInt64, Array: false, Nullable: true}},
			{val: &vuint, attr: Attr{Type: AttrTypeUint, Array: false, Nullable: true}},
			{val: &vuint8, attr: Attr{Type: AttrTypeUint8, Array: false, Nullable: true}},
			{val: &vuint16, attr: Attr{Type: AttrTypeUint16, Array: false, Nullable: true}},
			{val: &vuint32, attr: Attr{Type: AttrTypeUint32, Array: false, Nullable: true}},
			{val: &vuint64, attr: Attr{Type: AttrTypeUint64, Array: false, Nullable: true}},
			{val: &vfloat32, attr: Attr{Type: AttrTypeFloat32, Array: false, Nullable: true}},
			{val: &vfloat64, attr: Attr{Type: AttrTypeFloat64, Array: false, Nullable: true}},
			{val: &vbool, attr: Attr{Type: AttrTypeBool, Array: false, Nullable: true}},
			{val: &vtime, attr: Attr{Type: AttrTypeTime, Array: false, Nullable: true}},
		},
		"nullable array": {
			{val: &vstrarr, attr: Attr{Type: AttrTypeString, Array: true, Nullable: true}},
			{val: &vintarr, attr: Attr{Type: AttrTypeInt, Array: true, Nullable: true}},
			{val: &vint8arr, attr: Attr{Type: AttrTypeInt8, Array: true, Nullable: true}},
			{val: &vint16arr, attr: Attr{Type: AttrTypeInt16, Array: true, Nullable: true}},
			{val: &vint32arr, attr: Attr{Type: AttrTypeInt32, Array: true, Nullable: true}},
			{val: &vint64arr, attr: Attr{Type: AttrTypeInt64, Array: true, Nullable: true}},
			{val: &vuintarr, attr: Attr{Type: AttrTypeUint, Array: true, Nullable: true}},
			{val: &vuint8arr, attr: Attr{Type: AttrTypeUint8, Array: true, Nullable: true}},
			{val: &vuint16arr, attr: Attr{Type: AttrTypeUint16, Array: true, Nullable: true}},
			{val: &vuint32arr, attr: Attr{Type: AttrTypeUint32, Array: true, Nullable: true}},
			{val: &vuint64arr, attr: Attr{Type: AttrTypeUint64, Array: true, Nullable: true}},
			{val: &vfloat32arr, attr: Attr{Type: AttrTypeFloat32, Array: true, Nullable: true}},
			{val: &vfloat64arr, attr: Attr{Type: AttrTypeFloat64, Array: true, Nullable: true}},
			{val: &vboolarr, attr: Attr{Type: AttrTypeBool, Array: true, Nullable: true}},
			{val: &vtimearr, attr: Attr{Type: AttrTypeTime, Array: true, Nullable: true}},
		},
		"bytes": {
			{
				val:  []uint8{1, 2, 3, 4},
				attr: Attr{Type: AttrTypeBytes, Array: false, Nullable: false},
			},
			{
				val:  []uint8{4, 3, 2, 1},
				attr: Attr{Type: AttrTypeBytes, Array: true, Nullable: false},
			},
			{
				val:  &[]uint8{1, 2, 3, 4},
				attr: Attr{Type: AttrTypeBytes, Array: false, Nullable: true},
			},
			{
				val:  &[]uint8{4, 3, 2, 1},
				attr: Attr{Type: AttrTypeBytes, Array: true, Nullable: true},
			},
		},
	}

	for n, tests := range testData {
		t.Run(n, func(t *testing.T) {
			for _, test := range tests {
				p, _ := json.Marshal(test.val)
				v, err := UnmarshalToType(p, test.attr)

				assert.NoError(t, err)
				assert.Equal(t, test.val, v)
			}
		})
	}

	// Testing invalid attribute types
	p, _ := json.Marshal("12345")
	_, err := UnmarshalToType(p, Attr{Type: AttrTypeInvalid})
	assert.Error(t, err)

	p, _ = json.Marshal(nil)
	_, err = UnmarshalToType(p, Attr{Type: 9999999})
	assert.Error(t, err)

	_, err = UnmarshalToType([]byte("null"), Attr{Type: AttrTypeString})
	assert.Error(t, err)

	_, err = UnmarshalToType([]byte("falsch"), Attr{Type: AttrTypeBool})
	assert.Error(t, err)

	assert.Panics(t, func() {
		_, _ = UnmarshalToType([]byte("123"), Attr{Type: AttrTypeBytes})
	})
}

func TestRegisterAttrType(t *testing.T) {
	assert.Panics(t, func() {
		RegisterAttrType(AttrTypeInvalid, "test", nil, nil)
	})
}

func TestGetAttrTypeName(t *testing.T) {
	name, err := GetAttrTypeName(AttrTypeString, true, false)
	assert.NoError(t, err)
	assert.Equal(t, "string[]", name)

	name, err = GetAttrTypeName(AttrTypeString, false, true)
	assert.NoError(t, err)
	assert.Equal(t, "string (nullable)", name)

	name, err = GetAttrTypeName(AttrTypeString, true, true)
	assert.NoError(t, err)
	assert.Equal(t, "string[] (nullable)", name)

	_, err = GetAttrTypeName(9999, false, true)
	assert.Error(t, err)

	SetAttrTypeNameFunc(nil)

	name, err = GetAttrTypeName(AttrTypeString, true, true)
	assert.NoError(t, err)
	assert.Equal(t, "string", name)

	SetAttrTypeNameFunc(goLikeNameFunc)

	name, err = GetAttrTypeName(AttrTypeString, true, true)
	assert.NoError(t, err)
	assert.Equal(t, "*[]string", name)

	SetAttrTypeNameFunc(DefaultNameFunc)
}

func goLikeNameFunc(name string, array, nullable bool) string {
	if array {
		name = "[]" + name
	}

	if nullable {
		name = "*" + name
	}

	return name
}
