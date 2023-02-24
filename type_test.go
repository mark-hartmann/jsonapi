package jsonapi_test

import (
	. "github.com/mark-hartmann/jsonapi"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_AddAttr(t *testing.T) {
	attrTests := map[string]struct {
		attr Attr
		err  bool
	}{
		"attr string": {
			attr: Attr{
				Name:     "attr1",
				Type:     AttrTypeString,
				Nullable: false,
			},
		},
		"attr *string": {
			attr: Attr{
				Name:     "attr",
				Type:     AttrTypeString,
				Nullable: true,
			},
		},
		"attr *[]string": {
			attr: Attr{
				Name:     "attr",
				Type:     AttrTypeString,
				Array:    true,
				Nullable: true,
			},
		},
		// AttrTypeBytes is implicitly Array=true
		"attr bytes (non-array)": {
			attr: Attr{
				Name:     "attr",
				Type:     AttrTypeBytes,
				Nullable: true,
			},
		},
		"attr (invalid type)": {
			attr: Attr{Name: "invalid"},
			err:  true,
		},
		"attr (no name)": {
			attr: Attr{Type: AttrTypeBool},
			err:  true,
		},
		"attr (AttrTypeInvalid)": {
			attr: Attr{Type: AttrTypeInvalid},
			err:  true,
		},
		"attr (illegal name relationships)": {
			attr: Attr{Name: "relationships"},
			err:  true,
		},
		"attr (illegal name links)": {
			attr: Attr{Name: "links"},
			err:  true,
		},
		"attr (illegal name type)": {
			attr: Attr{Name: "type"},
			err:  true,
		},
		"attr (illegal name id)": {
			attr: Attr{Name: "id"},
			err:  true,
		},
	}

	for name, test := range attrTests {
		t.Run(name, func(t *testing.T) {
			typ := &Type{
				Name: "type",
			}

			err := typ.AddAttr(test.attr)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	typ := &Type{
		Name: "type1",
	}
	_ = typ.AddAttr(Attr{
		Name: "attr1",
		Type: AttrTypeString,
	})

	// Add invalid attribute (name already used)
	err := typ.AddAttr(Attr{Name: "attr1", Type: AttrTypeString})
	assert.Error(t, err)

	err = typ.AddAttr(Attr{Name: "some-attr", Type: 9999})
	assert.Error(t, err)
}

func TestType_AddRel(t *testing.T) {
	relTests := map[string]struct {
		rel Rel
		err bool
	}{
		"rel": {
			rel: Rel{
				FromName: "rel1",
				ToType:   "type1",
			},
		},
		"invalid rel (no name)": {
			rel: Rel{},
			err: true,
		},
		"invalid rel (illegal name id)": {
			rel: Rel{FromName: "id"},
			err: true,
		},
		"invalid rel (illegal name type)": {
			rel: Rel{FromName: "type"},
			err: true,
		},
		"invalid rel (empty type)": {
			rel: Rel{FromName: "invalid"},
			err: true,
		},
	}

	for name, test := range relTests {
		t.Run(name, func(t *testing.T) {
			typ := &Type{
				Name: "type",
			}

			err := typ.AddRel(test.rel)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	typ := &Type{
		Name: "type1",
	}
	_ = typ.AddRel(Rel{
		FromName: "rel1",
		ToType:   "type1",
	})

	// Add invalid relationship (name already used)
	err := typ.AddRel(Rel{FromName: "rel1", ToType: "type1"})
	assert.Error(t, err)
}

// TODO Add tests with attributes and relationships.
func TestTypeEqual(t *testing.T) {
	assert := assert.New(t)

	// Two empty types
	typ1 := Type{}
	typ2 := Type{}
	assert.True(typ1.Equal(typ2))

	typ1.Name = "type1"
	typ2.Name = "type1"
	assert.True(typ1.Equal(typ2))

	typ1.Name = "type1"
	typ2.Name = "type2"
	assert.False(typ1.Equal(typ2))

	// Make sure NewFunc is ignored.
	typ1.Name = "type1"
	typ1.NewFunc = func() Resource {
		return nil
	}
	typ2.Name = "type1"
	typ2.NewFunc = func() Resource {
		return &SoftResource{}
	}
	assert.True(typ1.Equal(typ2))
}

func TestTypeNewFunc(t *testing.T) {
	assert := assert.New(t)

	// NewFunc is nil
	typ := &Type{}
	assert.Equal(&SoftResource{Type: typ}, typ.New())

	// NewFunc is not nil
	typ = &Type{
		NewFunc: func() Resource {
			res := &SoftResource{}
			res.SetID("abc123")
			return res
		},
	}
	assert.Equal("abc123", typ.New().Get("id").(string))
}

//func TestAttrUnmarshalToType(t *testing.T) {
//	assert := assert.New(t)
//
//	var (
//		vstr     = "str"
//		vint     = int(1)
//		vint8    = int8(8)
//		vint16   = int16(16)
//		vint32   = int32(32)
//		vint64   = int64(64)
//		vuint    = uint(1)
//		vuint8   = uint8(8)
//		vuint16  = uint16(16)
//		vuint32  = uint32(32)
//		vuint64  = uint64(64)
//		vfloat32 = float32(math.MaxFloat32)
//		vfloat64 = math.MaxFloat64
//		vbool    = true
//
//		vstrarr     = []string{"str"}
//		vintarr     = []int{1}
//		vint8arr    = []int8{8}
//		vint16arr   = []int16{16}
//		vint32arr   = []int32{32}
//		vint64arr   = []int64{64}
//		vuintarr    = []uint{1}
//		vuint8arr   = []uint8{8}
//		vuint16arr  = []uint16{16}
//		vuint32arr  = []uint32{32}
//		vuint64arr  = []uint64{64}
//		vfloat32arr = []float32{math.MaxFloat32}
//		vfloat64arr = []float64{math.MaxFloat64}
//		vboolarr    = []bool{true}
//		vtimearr    = []time.Time{{}}
//	)
//
//	tests := []struct {
//		val interface{}
//	}{
//		{val: "str"},       // string
//		{val: 1},           // int
//		{val: int8(8)},     // int8
//		{val: int16(16)},   // int16
//		{val: int32(32)},   // int32
//		{val: int64(64)},   // int64
//		{val: uint(1)},     // uint
//		{val: uint8(8)},    // uint8
//		{val: uint16(16)},  // uint16
//		{val: uint32(32)},  // uint32
//		{val: uint64(64)},  // uint64
//		{val: true},        // bool
//		{val: time.Time{}}, // time
//
//		{val: &vstr},        // *string
//		{val: &vint},        // *int
//		{val: &vint8},       // *int8
//		{val: &vint16},      // *int16
//		{val: &vint32},      // *int32
//		{val: &vint64},      // *int64
//		{val: &vuint},       // *uint
//		{val: &vuint8},      // *uint8
//		{val: &vuint16},     // *uint16
//		{val: &vuint32},     // *uint32
//		{val: &vuint64},     // *uint64
//		{val: &vfloat32},    // *float32
//		{val: &vfloat64},    // *float64
//		{val: &vbool},       // *bool
//		{val: &time.Time{}}, // *time
//
//		{val: vstrarr},     // []string
//		{val: vintarr},     // []int
//		{val: vint8arr},    // []int8
//		{val: vint16arr},   // []int16
//		{val: vint32arr},   // []int32
//		{val: vint64arr},   // []int64
//		{val: vuintarr},    // []uint
//		{val: vuint8arr},   // []uint8
//		{val: vuint16arr},  // []uint16
//		{val: vuint32arr},  // []uint32
//		{val: vuint64arr},  // []uint64
//		{val: vfloat32arr}, // []float32
//		{val: vfloat64arr}, // []float64
//		{val: vboolarr},    // []bool
//		{val: vtimearr},    // []time.Time
//
//		{val: &vstrarr},     // *[]string
//		{val: &vintarr},     // *[]int
//		{val: &vint8arr},    // *[]int8
//		{val: &vint16arr},   // *[]int16
//		{val: &vint32arr},   // *[]int32
//		{val: &vint64arr},   // *[]int64
//		{val: &vuintarr},    // *[]uint
//		{val: &vuint8arr},   // *[]uint8
//		{val: &vuint16arr},  // *[]uint16
//		{val: &vuint32arr},  // *[]uint32
//		{val: &vuint64arr},  // *[]uint64
//		{val: &vfloat32arr}, // *[]float32
//		{val: &vfloat64arr}, // *[]float64
//		{val: &vboolarr},    // *[]bool
//		{val: &vtimearr},    // *[]time.Time
//	}
//
//	attr := Attr{}
//
//	for _, test := range tests {
//		attr.Type, attr.Array, attr.Nullable = GetAttrType(fmt.Sprintf("%T", test.val))
//		p, _ := json.Marshal(test.val)
//		val, err := attr.UnmarshalToType(p)
//		assert.NoError(err)
//		assert.Equal(test.val, val)
//		assert.Equal(fmt.Sprintf("%T", test.val), fmt.Sprintf("%T", val))
//	}
//
//	// boolean not-true value
//	attr.Array = false
//	attr.Nullable = false
//	attr.Type = AttrTypeBool
//	val, err := attr.UnmarshalToType([]byte("nottrue"))
//	assert.Error(err)
//	assert.Nil(val)
//
//	// Invalid attribute type
//	attr.Type = AttrTypeInvalid
//	val, err = attr.UnmarshalToType([]byte("invalid"))
//	assert.Error(err)
//	assert.Nil(val)
//}
//
//func TestAttrUnmarshalToType_Nil(t *testing.T) {
//	attr := Attr{Type: AttrTypeString}
//	v, err := attr.UnmarshalToType(nil)
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeString, Nullable: true}
//	v, err = attr.UnmarshalToType(nil)
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeString, Array: true}
//	v, err = attr.UnmarshalToType(nil)
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeString, Nullable: true, Array: true}
//	v, err = attr.UnmarshalToType(nil)
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	// The compiler is smart enough, but better safe than sorry.
//	attr = Attr{Type: AttrTypeString, Nullable: true, Array: true}
//	v, err = attr.UnmarshalToType(([]byte)(nil))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//}
//
//func TestAttrUnmarshalToType_Null(t *testing.T) {
//	attr := Attr{Type: AttrTypeString}
//	v, err := attr.UnmarshalToType([]byte("null"))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeString, Nullable: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.NoError(t, err)
//	assert.Equal(t, (*string)(nil), v)
//
//	attr = Attr{Type: AttrTypeString, Array: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeString, Nullable: true, Array: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.NoError(t, err)
//	assert.Equal(t, (*[]string)(nil), v)
//
//	attr = Attr{Type: AttrTypeUint8}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeUint8, Nullable: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.NoError(t, err)
//	assert.Equal(t, (*uint8)(nil), v)
//
//	attr = Attr{Type: AttrTypeUint8, Array: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeUint8, Nullable: true, Array: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.NoError(t, err)
//	assert.Equal(t, (*[]uint8)(nil), v)
//
//	// AttrTypeBytes is implicitly Array=true
//	attr = Attr{Type: AttrTypeBytes}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeBytes, Array: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.Error(t, err)
//	assert.Equal(t, nil, v)
//
//	attr = Attr{Type: AttrTypeBytes, Nullable: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.NoError(t, err)
//	assert.Equal(t, (*[]uint8)(nil), v)
//
//	attr = Attr{Type: AttrTypeBytes, Nullable: true, Array: true}
//	v, err = attr.UnmarshalToType([]byte("null"))
//	assert.NoError(t, err)
//	assert.Equal(t, (*[]uint8)(nil), v)
//}
//
//func TestAttrUnmarshalToType_Bytes(t *testing.T) {
//	t.Run("bytes", func(t *testing.T) {
//		bytes := []byte("hello world")
//		attr := Attr{Name: "bytes", Type: AttrTypeBytes, Array: true}
//		p, _ := json.Marshal(bytes)
//		assert.Equal(t, `"aGVsbG8gd29ybGQ="`, string(p))
//
//		val, err := attr.UnmarshalToType(p)
//		assert.NoError(t, err)
//		assert.Equal(t, bytes, val)
//	})
//
//	t.Run("nullable bytes", func(t *testing.T) {
//		bytes := []byte("hello world")
//		attr := Attr{Name: "bytes", Type: AttrTypeBytes, Array: true, Nullable: true}
//		p, _ := json.Marshal(bytes)
//		assert.Equal(t, `"aGVsbG8gd29ybGQ="`, string(p))
//
//		val, err := attr.UnmarshalToType(p)
//		assert.NoError(t, err)
//		assert.Equal(t, &bytes, val)
//	})
//
//	t.Run("[]uint8", func(t *testing.T) {
//		bytes := []uint8{1, 2, 4, 8, 16, 32}
//		attr := Attr{Name: "uint8arr", Type: AttrTypeUint8, Array: true, Nullable: false}
//
//		val, err := attr.UnmarshalToType([]byte("[1,2,4,8,16,32]"))
//		assert.NoError(t, err)
//		assert.Equal(t, bytes, val)
//
//		val, err = attr.UnmarshalToType([]byte("[]"))
//		assert.NoError(t, err)
//		assert.Equal(t, []byte{}, val)
//	})
//
//	t.Run("nullable []uint8", func(t *testing.T) {
//		attr := Attr{Name: "uint8arr", Type: AttrTypeUint8, Array: true, Nullable: true}
//
//		val, err := attr.UnmarshalToType([]byte("[1,2,4,8,16,32]"))
//		assert.NoError(t, err)
//		assert.Equal(t, &[]uint8{1, 2, 4, 8, 16, 32}, val)
//
//		val, err = attr.UnmarshalToType([]byte("[]"))
//		assert.NoError(t, err)
//		assert.Equal(t, &[]byte{}, val)
//	})
//
//	// Invalid slide of bytes
//	attr := Attr{Type: AttrTypeBytes}
//
//	assert.Panics(t, func() {
//		_, _ = attr.UnmarshalToType([]byte("invalid"))
//	})
//}

func TestRelInvert(t *testing.T) {
	assert := assert.New(t)

	rel := Rel{
		FromName: "rel1",
		FromType: "type1",
		ToOne:    true,
		ToName:   "rel2",
		ToType:   "type2",
		FromOne:  false,
	}

	invRel := rel.Invert()

	assert.Equal("rel2", invRel.FromName)
	assert.Equal("type1", invRel.ToType)
	assert.Equal(false, invRel.ToOne)
	assert.Equal("rel1", invRel.ToName)
	assert.Equal("type2", invRel.FromType)
	assert.Equal(true, invRel.FromOne)
}

func TestRelNormalize(t *testing.T) {
	assert := assert.New(t)

	rel := Rel{
		FromName: "rel2",
		FromType: "type2",
		ToOne:    false,
		ToName:   "rel1",
		ToType:   "type1",
		FromOne:  true,
	}

	// Normalize should return the inverse because
	// type1 comes before type2 alphabetically.
	norm := rel.Normalize()
	assert.Equal("type1", norm.FromType)
	assert.Equal("rel1", norm.FromName)
	assert.Equal(true, norm.ToOne)
	assert.Equal("type2", norm.ToType)
	assert.Equal("rel2", norm.ToName)
	assert.Equal(false, norm.FromOne)

	// Normalize again, but it should stay the same.
	norm = norm.Normalize()
	assert.Equal("type1", norm.FromType)
	assert.Equal("rel1", norm.FromName)
	assert.Equal(true, norm.ToOne)
	assert.Equal("type2", norm.ToType)
	assert.Equal("rel2", norm.ToName)
	assert.Equal(false, norm.FromOne)
}

func TestRelString(t *testing.T) {
	assert := assert.New(t)

	rel := Rel{
		FromName: "rel2",
		FromType: "type2",
		ToOne:    false,
		ToName:   "rel1",
		ToType:   "type1",
		FromOne:  true,
	}

	assert.Equal("type1_rel1_type2_rel2", rel.String())
	assert.Equal("type1_rel1_type2_rel2", rel.Invert().String())
}

func TestGetAttrType(t *testing.T) {
	testData := []struct {
		str      string
		typ      int
		array    bool
		nullable bool
	}{
		{
			str:      "string",
			typ:      AttrTypeString,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]string",
			typ:      AttrTypeString,
			array:    true,
			nullable: false,
		},
		{
			str:      "*string",
			typ:      AttrTypeString,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]string",
			typ:      AttrTypeString,
			array:    true,
			nullable: true,
		},
		{
			str:      "int",
			typ:      AttrTypeInt,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]int",
			typ:      AttrTypeInt,
			array:    true,
			nullable: false,
		},
		{
			str:      "*int",
			typ:      AttrTypeInt,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]int",
			typ:      AttrTypeInt,
			array:    true,
			nullable: true,
		},
		{
			str:      "int8",
			typ:      AttrTypeInt8,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]int8",
			typ:      AttrTypeInt8,
			array:    true,
			nullable: false,
		},
		{
			str:      "*int8",
			typ:      AttrTypeInt8,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]int8",
			typ:      AttrTypeInt8,
			array:    true,
			nullable: true,
		},
		{
			str:      "int16",
			typ:      AttrTypeInt16,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]int16",
			typ:      AttrTypeInt16,
			array:    true,
			nullable: false,
		},
		{
			str:      "*int16",
			typ:      AttrTypeInt16,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]int16",
			typ:      AttrTypeInt16,
			array:    true,
			nullable: true,
		},
		{
			str:      "int32",
			typ:      AttrTypeInt32,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]int32",
			typ:      AttrTypeInt32,
			array:    true,
			nullable: false,
		},
		{
			str:      "*int32",
			typ:      AttrTypeInt32,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]int32",
			typ:      AttrTypeInt32,
			array:    true,
			nullable: true,
		},
		{
			str:      "int64",
			typ:      AttrTypeInt64,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]int64",
			typ:      AttrTypeInt64,
			array:    true,
			nullable: false,
		},
		{
			str:      "*int64",
			typ:      AttrTypeInt64,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]int64",
			typ:      AttrTypeInt64,
			array:    true,
			nullable: true,
		},
		{
			str:      "float32",
			typ:      AttrTypeFloat32,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]float32",
			typ:      AttrTypeFloat32,
			array:    true,
			nullable: false,
		},
		{
			str:      "*float32",
			typ:      AttrTypeFloat32,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]float32",
			typ:      AttrTypeFloat32,
			array:    true,
			nullable: true,
		},
		{
			str:      "float64",
			typ:      AttrTypeFloat64,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]float64",
			typ:      AttrTypeFloat64,
			array:    true,
			nullable: false,
		},
		{
			str:      "*float64",
			typ:      AttrTypeFloat64,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]float64",
			typ:      AttrTypeFloat64,
			array:    true,
			nullable: true,
		},
		{
			str:      "uint",
			typ:      AttrTypeUint,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]uint",
			typ:      AttrTypeUint,
			array:    true,
			nullable: false,
		},
		{
			str:      "*uint",
			typ:      AttrTypeUint,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]uint",
			typ:      AttrTypeUint,
			array:    true,
			nullable: true,
		},
		{
			str:      "uint8",
			typ:      AttrTypeUint8,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]uint8",
			typ:      AttrTypeUint8,
			array:    true,
			nullable: false,
		},
		{
			str:      "*uint8",
			typ:      AttrTypeUint8,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]uint8",
			typ:      AttrTypeUint8,
			array:    true,
			nullable: true,
		},
		{
			str:      "uint16",
			typ:      AttrTypeUint16,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]uint16",
			typ:      AttrTypeUint16,
			array:    true,
			nullable: false,
		},
		{
			str:      "*uint16",
			typ:      AttrTypeUint16,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]uint16",
			typ:      AttrTypeUint16,
			array:    true,
			nullable: true,
		},
		{
			str:      "uint32",
			typ:      AttrTypeUint32,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]uint32",
			typ:      AttrTypeUint32,
			array:    true,
			nullable: false,
		},
		{
			str:      "*uint32",
			typ:      AttrTypeUint32,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]uint32",
			typ:      AttrTypeUint32,
			array:    true,
			nullable: true,
		},
		{
			str:      "uint64",
			typ:      AttrTypeUint64,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]uint64",
			typ:      AttrTypeUint64,
			array:    true,
			nullable: false,
		},
		{
			str:      "*uint64",
			typ:      AttrTypeUint64,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]uint64",
			typ:      AttrTypeUint64,
			array:    true,
			nullable: true,
		},
		{
			str:      "bool",
			typ:      AttrTypeBool,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]bool",
			typ:      AttrTypeBool,
			array:    true,
			nullable: false,
		},
		{
			str:      "*bool",
			typ:      AttrTypeBool,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]bool",
			typ:      AttrTypeBool,
			array:    true,
			nullable: true,
		},
		{
			str:      "time.Time",
			typ:      AttrTypeTime,
			array:    false,
			nullable: false,
		},
		{
			str:      "[]time.Time",
			typ:      AttrTypeTime,
			array:    true,
			nullable: false,
		},
		{
			str:      "*time.Time",
			typ:      AttrTypeTime,
			array:    false,
			nullable: true,
		},
		{
			str:      "*[]time.Time",
			typ:      AttrTypeTime,
			array:    true,
			nullable: true,
		},
		{
			str:      "invalid",
			typ:      AttrTypeInvalid,
			array:    false,
			nullable: false,
		},
		{
			str:      "",
			typ:      AttrTypeInvalid,
			array:    false,
			nullable: false,
		},
		{
			str: "byte",
			typ: AttrTypeUint8,
		},
		{
			str:   "[]byte",
			typ:   AttrTypeUint8,
			array: true,
		},
		{
			str:      "*[]byte",
			typ:      AttrTypeUint8,
			array:    true,
			nullable: true,
		},
	}

	for _, test := range testData {
		t.Run(test.str, func(t *testing.T) {
			typ, array, nullable := GetAttrType(test.str)
			assert.Equal(t, test.typ, typ)
			assert.Equal(t, test.nullable, nullable)
			assert.Equal(t, test.array, array)
		})
	}
}

//func TestGetAttrTypeString(t *testing.T) {
//	assert := assert.New(t)
//
//	assert.Equal("string", Types.GetName(AttrTypeString, false, false))
//	assert.Equal("int", Types.GetName(AttrTypeInt, false, false))
//	assert.Equal("int8", Types.GetName(AttrTypeInt8, false, false))
//	assert.Equal("int16", Types.GetName(AttrTypeInt16, false, false))
//	assert.Equal("int32", Types.GetName(AttrTypeInt32, false, false))
//	assert.Equal("int64", Types.GetName(AttrTypeInt64, false, false))
//	assert.Equal("uint", Types.GetName(AttrTypeUint, false, false))
//	assert.Equal("uint8", Types.GetName(AttrTypeUint8, false, false))
//	assert.Equal("uint16", Types.GetName(AttrTypeUint16, false, false))
//	assert.Equal("uint32", Types.GetName(AttrTypeUint32, false, false))
//	assert.Equal("uint64", Types.GetName(AttrTypeUint64, false, false))
//	assert.Equal("float32", Types.GetName(AttrTypeFloat32, false, false))
//	assert.Equal("float64", Types.GetName(AttrTypeFloat64, false, false))
//	assert.Equal("bool", Types.GetName(AttrTypeBool, false, false))
//	assert.Equal("time.Time", Types.GetName(AttrTypeTime, false, false))
//
//	assert.Equal("*string", Types.GetName(AttrTypeString, false, true))
//	assert.Equal("*int", Types.GetName(AttrTypeInt, false, true))
//	assert.Equal("*int8", Types.GetName(AttrTypeInt8, false, true))
//	assert.Equal("*int16", Types.GetName(AttrTypeInt16, false, true))
//	assert.Equal("*int32", Types.GetName(AttrTypeInt32, false, true))
//	assert.Equal("*int64", Types.GetName(AttrTypeInt64, false, true))
//	assert.Equal("*uint", Types.GetName(AttrTypeUint, false, true))
//	assert.Equal("*uint8", Types.GetName(AttrTypeUint8, false, true))
//	assert.Equal("*uint16", Types.GetName(AttrTypeUint16, false, true))
//	assert.Equal("*uint32", Types.GetName(AttrTypeUint32, false, true))
//	assert.Equal("*uint64", Types.GetName(AttrTypeUint64, false, true))
//	assert.Equal("*float32", Types.GetName(AttrTypeFloat32, false, true))
//	assert.Equal("*float64", Types.GetName(AttrTypeFloat64, false, true))
//	assert.Equal("*bool", Types.GetName(AttrTypeBool, false, true))
//	assert.Equal("*time.Time", Types.GetName(AttrTypeTime, false, true))
//
//	assert.Equal("[]string", Types.GetName(AttrTypeString, true, false))
//	assert.Equal("[]int", Types.GetName(AttrTypeInt, true, false))
//	assert.Equal("[]int8", Types.GetName(AttrTypeInt8, true, false))
//	assert.Equal("[]int16", Types.GetName(AttrTypeInt16, true, false))
//	assert.Equal("[]int32", Types.GetName(AttrTypeInt32, true, false))
//	assert.Equal("[]int64", Types.GetName(AttrTypeInt64, true, false))
//	assert.Equal("[]uint", Types.GetName(AttrTypeUint, true, false))
//	assert.Equal("[]uint8", Types.GetName(AttrTypeUint8, true, false))
//	assert.Equal("[]uint16", Types.GetName(AttrTypeUint16, true, false))
//	assert.Equal("[]uint32", Types.GetName(AttrTypeUint32, true, false))
//	assert.Equal("[]uint64", Types.GetName(AttrTypeUint64, true, false))
//	assert.Equal("[]float32", Types.GetName(AttrTypeFloat32, true, false))
//	assert.Equal("[]float64", Types.GetName(AttrTypeFloat64, true, false))
//	assert.Equal("[]bool", Types.GetName(AttrTypeBool, true, false))
//	assert.Equal("[]time.Time", Types.GetName(AttrTypeTime, true, false))
//
//	assert.Equal("*[]string", Types.GetName(AttrTypeString, true, true))
//	assert.Equal("*[]int", Types.GetName(AttrTypeInt, true, true))
//	assert.Equal("*[]int8", Types.GetName(AttrTypeInt8, true, true))
//	assert.Equal("*[]int16", Types.GetName(AttrTypeInt16, true, true))
//	assert.Equal("*[]int32", Types.GetName(AttrTypeInt32, true, true))
//	assert.Equal("*[]int64", Types.GetName(AttrTypeInt64, true, true))
//	assert.Equal("*[]uint", Types.GetName(AttrTypeUint, true, true))
//	assert.Equal("*[]uint8", Types.GetName(AttrTypeUint8, true, true))
//	assert.Equal("*[]uint16", Types.GetName(AttrTypeUint16, true, true))
//	assert.Equal("*[]uint32", Types.GetName(AttrTypeUint32, true, true))
//	assert.Equal("*[]uint64", Types.GetName(AttrTypeUint64, true, true))
//	assert.Equal("*[]float32", Types.GetName(AttrTypeFloat32, true, true))
//	assert.Equal("*[]float64", Types.GetName(AttrTypeFloat64, true, true))
//	assert.Equal("*[]bool", Types.GetName(AttrTypeBool, true, true))
//	assert.Equal("*[]time.Time", Types.GetName(AttrTypeTime, true, true))
//
//	assert.Equal("[]uint8", Types.GetName(AttrTypeBytes, true, false))
//	assert.Equal("[]uint8", Types.GetName(AttrTypeBytes, false, false))
//	assert.Equal("*[]uint8", Types.GetName(AttrTypeBytes, true, true))
//	assert.Equal("*[]uint8", Types.GetName(AttrTypeBytes, false, true))
//
//	assert.Equal("", Types.GetName(AttrTypeInvalid, false, false))
//	assert.Equal("", Types.GetName(999, false, false))
//}
//

func TestCopyType(t *testing.T) {
	assert := assert.New(t)

	typ1 := Type{
		Name: "type1",
		Attrs: map[string]Attr{
			"attr1": {
				Name:     "attr1",
				Type:     AttrTypeString,
				Nullable: true,
			},
		},
		Rels: map[string]Rel{
			"rel1": {
				FromName: "rel1",
				FromType: "type1",
				ToOne:    true,
				ToName:   "rel2",
				ToType:   "type2",
				FromOne:  false,
			},
		},
	}

	// Copy
	typ2 := typ1.Copy()

	assert.Equal("type1", typ2.Name)
	assert.Len(typ2.Attrs, 1)
	assert.Equal("attr1", typ2.Attrs["attr1"].Name)
	assert.Equal(AttrTypeString, typ2.Attrs["attr1"].Type)
	assert.True(typ2.Attrs["attr1"].Nullable)
	assert.Len(typ2.Rels, 1)
	assert.Equal("rel1", typ2.Rels["rel1"].FromName)
	assert.Equal("type2", typ2.Rels["rel1"].ToType)
	assert.True(typ2.Rels["rel1"].ToOne)
	assert.Equal("rel2", typ2.Rels["rel1"].ToName)
	assert.Equal("type1", typ2.Rels["rel1"].FromType)
	assert.False(typ2.Rels["rel1"].FromOne)

	// Modify original (copy should not change)
	typ1.Name = "type3"
	typ1.Attrs["attr2"] = Attr{
		Name: "attr2",
		Type: AttrTypeInt,
	}

	assert.Equal("type1", typ2.Name)
	assert.Len(typ2.Attrs, 1)

	typ1.Name = "type1"
	delete(typ1.Attrs, "attr2")

	// Modify copy (original should not change)
	typ2.Name = "type3"
	typ2.Attrs["attr2"] = Attr{
		Name: "attr2",
		Type: AttrTypeInt,
	}

	assert.Equal("type1", typ1.Name)
	assert.Len(typ1.Attrs, 1)
}
