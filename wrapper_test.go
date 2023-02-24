package jsonapi_test

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

var _ Resource = (*Wrapper)(nil)
var _ Copier = (*Wrapper)(nil)

func TestWrap(t *testing.T) {
	assert := assert.New(t)

	assert.Panics(func() {
		_ = Wrap("just a string")
	}, "panic when not a pointer to a struct")

	assert.Panics(func() {
		str := "just a string"
		_ = Wrap(&str)
	}, "panic when not a pointer to a struct")

	assert.NotPanics(func() {
		_ = Wrap(&mocktype{})
	}, "don't panic when a valid struct")

	assert.NotPanics(func() {
		_ = Wrap(mocktype{})
	}, "don't panic when a pointer to a valid struct")

	assert.Panics(func() {
		s := time.Now()
		_ = Wrap(&s)
	}, "panic when not a valid struct")

	assert.Panics(func() {
		s := struct {
			ID   string  `json:"id" api:"test-struct"`
			Test big.Int `json:"test" api:"attr,some-unknown-type"`
		}{}

		_ = Wrap(s)
	})

	// Empty tag == no tag
	assert.NotPanics(func() {
		s := struct {
			ID   string `json:"id" api:"test-struct"`
			Test string `json:"test" api:""`
		}{}

		_ = Wrap(s)
	})
}

func TestWrapStruct(t *testing.T) {
	assert := assert.New(t)

	res1 := mockType1{
		ID:  "res123",
		Str: "a_string",
	}

	wrap1 := Wrap(res1)

	// ID, type, field
	id, typ := wrap1.IDAndType()
	assert.Equal(res1.ID, id, "id")
	assert.Equal("mocktypes1", typ, "type")
	assert.Equal(res1.Str, wrap1.Get("str"), "str field")

	// Modifying the wrapper does not modify
	// the original value.
	wrap1.SetID("another_id")
	id, _ = wrap1.IDAndType()
	assert.Equal("another_id", id, "type")

	wrap1.Set("str", "another_string")
	assert.Equal("another_string", wrap1.Get("str"), "str field")

	// Modifying the original value does
	// not modify the wrapper.
	res1.Str = "new_string"
	assert.NotEqual(res1.Str, wrap1.Get("str"), "str field")
	assert.Equal("another_string", wrap1.Get("str"), "str field")

	res2 := mockType6{
		ID:        "id1",
		Obj:       testObjType{Prop1: "abc"},
		ObjPtr:    &testObjType{Prop2: "def"},
		ObjArr:    []testObjType{{Prop1: "abc"}},
		ObjArrPtr: &[]testObjType{{Prop1: "def"}},
	}

	wrap2 := Wrap(res2)

	// ID, type, field
	id, typ = wrap2.IDAndType()
	assert.Equal(res2.ID, id, "id")
	assert.Equal("mocktype6", typ, "type")
	assert.Equal(res2.Str, wrap2.Get("str"), "str field")

	// Modifying the wrapper does not modify the original value.
	wrap2.SetID("another_id")

	id, _ = wrap1.IDAndType()
	assert.Equal("another_id", id, "type")

	wrap2.Set("obj", testObjType{Prop1: "xyz"})
	assert.Equal(testObjType{Prop1: "xyz"}, wrap2.Get("obj"), "obj field")
	assert.NotEqual(res2.Obj, wrap2.Get("obj"), "obj field")

	// Modifying the original value does not modify the wrapper.
	res2.ObjPtr = &testObjType{Prop1: "xyz"}
	assert.NotEqual(res2.ObjPtr, wrap2.Get("objPtr"), "objPtr field")
	assert.Equal(&testObjType{Prop2: "def"}, wrap2.Get("objPtr"), "objPtr field")
}

func TestWrapper(t *testing.T) {
	assert := assert.New(t)

	loc, _ := time.LoadLocation("")

	res1 := &mockType1{
		ID:     "res123",
		Str:    "a_string",
		Int:    2,
		Int8:   8,
		Int16:  16,
		Int32:  32,
		Int64:  64,
		Uint:   4,
		Uint8:  8,
		Uint16: 16,
		Uint32: 32,
		Uint64: 64,
		Bool:   true,
		Time:   time.Date(2017, 1, 2, 3, 4, 5, 6, loc),
	}

	wrap1 := Wrap(res1)

	// ID and type
	id, typ := wrap1.IDAndType()
	assert.Equal(res1.ID, id, "id")
	assert.Equal("mocktypes1", typ, "type")

	wrap1.SetID("another-id")
	assert.Equal(res1.ID, "another-id", "set id")

	// Get attributes
	attr := wrap1.Attr("str")
	assert.Equal(Attr{
		Name:     "str",
		Type:     AttrTypeString,
		Nullable: false,
	}, attr, "get attribute (str)")
	assert.Equal(Attr{}, wrap1.Attr("nonexistent"), "get non-existent attribute")

	// Get relationships
	rel := wrap1.Rel("to-one")
	assert.Equal(Rel{
		FromName: "to-one",
		ToType:   "mocktypes2",
		ToOne:    true,
		ToName:   "",
		FromType: "mocktypes1",
		FromOne:  false,
	}, rel, "get relationship (to-one)")
	assert.Equal(Rel{}, wrap1.Rel("nonexistent"), "get non-existent relationship")

	// Get values (attributes)
	v1 := reflect.ValueOf(res1).Elem()
	for i := 0; i < v1.NumField(); i++ {
		f := v1.Field(i)
		sf := v1.Type().Field(i)
		n := sf.Tag.Get("json")

		if sf.Tag.Get("api") == "attr" {
			assert.Equal(f.Interface(), wrap1.Get(n), "api tag")
		}
	}

	// Set values (attributes)
	wrap1.Set("str", "another_string")
	assert.Equal("another_string", wrap1.Get("str"), "set string attribute")
	wrap1.Set("int", 3)
	assert.Equal(3, wrap1.Get("int"), "set int attribute")

	aStr := "another_string_ptr"
	aInt := int(123)
	aInt8 := int8(88)
	aInt16 := int16(1616)
	aInt32 := int32(3232)
	aInt64 := int64(6464)
	aUint := uint(456)
	aUint8 := uint8(88)
	aUint16 := uint16(1616)
	aUint32 := uint32(3232)
	aUint64 := uint64(6464)
	aBool := false
	aTime := time.Date(2018, 2, 3, 4, 5, 6, 7, loc)

	// Set the values (attributes) after the wrapping
	res2 := &mockType2{
		ID:        "res123",
		StrPtr:    &aStr,
		IntPtr:    &aInt,
		Int8Ptr:   &aInt8,
		Int16Ptr:  &aInt16,
		Int32Ptr:  &aInt32,
		Int64Ptr:  &aInt64,
		UintPtr:   &aUint,
		Uint8Ptr:  &aUint8,
		Uint16Ptr: &aUint16,
		Uint32Ptr: &aUint32,
		Uint64Ptr: &aUint64,
		BoolPtr:   &aBool,
		TimePtr:   &aTime,
	}

	wrap2 := Wrap(res2)

	// ID and type
	id, typ = wrap2.IDAndType()
	assert.Equal(res2.ID, id, "id 2")
	assert.Equal("mocktypes2", typ, "type 2")

	// Get values (attributes)
	v2 := reflect.ValueOf(res2).Elem()
	for i := 0; i < v2.NumField(); i++ {
		f := v2.Field(i)
		sf := v2.Type().Field(i)
		n := sf.Tag.Get("json")

		if sf.Tag.Get("api") == "attr" {
			assert.Equal(f.Interface(), wrap2.Get(n), "api tag 2")
		}
	}

	// Set values (attributes)
	var (
		anotherString = "anotherString"
		newInt        = 3
	)

	wrap2.Set("strptr", &anotherString)
	assert.Equal(&anotherString, wrap2.Get("strptr"), "set string pointer attribute")

	wrap2.Set("intptr", &newInt)
	assert.Equal(&newInt, wrap2.Get("intptr"), "set int pointer attribute")

	wrap2.Set("uintptr", nil)
	assert.Equal((*uint)(nil), wrap2.Get("uintptr"))
	assert.NotEqual(nil, wrap2.Get("uintptr"))

	assert.Equal((*uint)(nil), res2.UintPtr)
	assert.NotEqual(nil, res2.UintPtr)

	// New
	wrap3 := wrap1.New()
	wrap3Type := wrap3.GetType()

	for _, attr := range wrap1.Attrs() {
		assert.Equal(wrap1.Attr(attr.Name), wrap3Type.Attrs[attr.Name], "copied attribute")
	}

	for _, rel := range wrap1.Rels() {
		assert.Equal(wrap1.Rel(rel.FromName), wrap3Type.Rels[rel.FromName], "copied relationship")
	}

	// Copy
	wrap3 = wrap1.Copy()

	for _, attr := range wrap1.Attrs() {
		assert.Equal(wrap1.Attr(attr.Name), wrap3Type.Attrs[attr.Name], "copied attribute")
	}

	for _, rel := range wrap1.Rels() {
		assert.Equal(wrap1.Rel(rel.FromName), wrap3Type.Rels[rel.FromName], "copied relationship")
	}

	wrap3.Set("str", "another string")
	assert.NotEqual(
		wrap1.Get("str"),
		wrap3.Get("str"),
		"modified value does not affect original",
	)

	res4 := &mockType4{
		BoolArr: []bool{true, false},
	}
	wrap4 := Wrap(res4)

	assert.Equal([]bool{true, false}, wrap4.Get("boolarr"))
	assert.Equal([]uint{}, wrap4.Get("uintarr"))

	res5 := &mockType5{
		BoolArrPtr: &[]bool{true, false},
	}
	wrap5 := Wrap(res5)

	assert.Equal(&[]bool{true, false}, res5.BoolArrPtr)
	assert.Equal(res5.BoolArrPtr, wrap5.Get("boolarrptr"))

	assert.Equal((*[]uint)(nil), res5.UintArrPtr)
	assert.Equal(res5.UintArrPtr, wrap5.Get("uintarrptr"))

	wrap5.Set("boolarrptr", nil)
	assert.Equal((*[]bool)(nil), res5.BoolArrPtr)
	assert.Equal(res5.BoolArrPtr, wrap5.Get("boolarrptr"))

	wrap5.Set("strarrptr", &[]string{"foo", "bar"})
	assert.Equal(&[]string{"foo", "bar"}, res5.StrArrPtr)
	assert.Equal(res5.StrArrPtr, wrap5.Get("strarrptr"))

	wrap5.Set("strarrptr", (*[]string)(nil))
	assert.Equal((*[]string)(nil), res5.StrArrPtr)
	assert.Equal(res5.StrArrPtr, wrap5.Get("strarrptr"))
}

func TestWrapperSet(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		typ   string // "1" for mockType1, "2" for mockType2
		field string
		val   interface{}
	}{
		{typ: "1", field: "str", val: "astring"},
		{typ: "1", field: "int", val: int(9)},
		{typ: "1", field: "int8", val: int8(9)},
		{typ: "1", field: "int16", val: int16(9)},
		{typ: "1", field: "int32", val: int32(9)},
		{typ: "1", field: "int64", val: int64(9)},
		{typ: "1", field: "uint", val: uint(9)},
		{typ: "1", field: "uint8", val: uint8(9)},
		{typ: "1", field: "uint16", val: uint16(9)},
		{typ: "1", field: "uint32", val: uint32(9)},
		{typ: "1", field: "uint64", val: uint64(9)},
		{typ: "1", field: "bool", val: bool(true)},
		{typ: "2", field: "strptr", val: ptr("astring")},
		{typ: "2", field: "intptr", val: ptr(int(9))},
		{typ: "2", field: "int8ptr", val: ptr(int8(9))},
		{typ: "2", field: "int16ptr", val: ptr(int16(9))},
		{typ: "2", field: "int32ptr", val: ptr(int32(9))},
		{typ: "2", field: "int64ptr", val: ptr(int64(9))},
		{typ: "2", field: "uintptr", val: ptr(uint(9))},
		{typ: "2", field: "uint8ptr", val: ptr(uint8(9))},
		{typ: "2", field: "uint16ptr", val: ptr(uint16(9))},
		{typ: "2", field: "uint32ptr", val: ptr(uint32(9))},
		{typ: "2", field: "uint64ptr", val: ptr(uint64(9))},
		{typ: "2", field: "boolptr", val: ptr(bool(true))},
	}

	for _, test := range tests {
		if test.typ == "1" {
			res1 := Wrap(&mockType1{})
			res1.Set(test.field, test.val)
			assert.EqualValues(test.val, res1.Get(test.field))
		}
	}

	t.Run("custom type unmarshaler", func(t *testing.T) {
		res := Wrap(&mockType6{})
		assert.NotNil(res)

		assert.Equal("", res.Get("str"))
		assert.Equal((*string)(nil), res.Get("strPtr"))
		assert.Equal([]string{}, res.Get("strArr"))
		assert.Equal([][]float32{}, res.Get("float32Matrix"))
		assert.Equal((*[]string)(nil), res.Get("strPtrArr"))

		assert.Equal(testObjType{}, res.Get("obj"))
		assert.Equal((*testObjType)(nil), res.Get("objPtr"))
		assert.Equal([]testObjType{}, res.Get("objArr"))
		assert.Equal((*[]testObjType)(nil), res.Get("objArrPtr"))

		obj := testObjType{Prop1: "foo", Prop2: "bar", Prop3: "baz"}
		objPtr := &testObjType{Prop1: "foo", Prop2: "bar", Prop3: "baz"}
		objArr := []testObjType{{Prop1: "foo", Prop2: "bar", Prop3: "baz"}}
		objArrPtr := &[]testObjType{{Prop1: "foo", Prop2: "bar", Prop3: "baz"}}

		res.Set("obj", obj)
		assert.Equal(obj, res.Get("obj"))

		res.Set("objPtr", objPtr)
		assert.Equal(objPtr, res.Get("objPtr"))

		res.Set("objArr", objArr)
		assert.Equal(objArr, res.Get("objArr"))

		res.Set("objArrPtr", objArrPtr)
		assert.Equal(objArrPtr, res.Get("objArrPtr"))
	})
}

func TestWrapperGetAndSetErrors(t *testing.T) {
	assert := assert.New(t)

	mt := &mocktype{}
	wrap := Wrap(mt)

	// Get on empty field name
	assert.Panics(func() {
		_ = wrap.Get("")
	})

	// Get on unknown field name
	assert.Panics(func() {
		_ = wrap.Get("unknown")
	})

	// Set on empty field name
	assert.Panics(func() {
		wrap.Set("", "")
	})

	// Set on unknown field name
	assert.Panics(func() {
		wrap.Set("unknown", "")
	})

	// Set with value of wrong type
	assert.Panics(func() {
		wrap.Set("str", 42)
	})
}

func TestReflectTypeUnmarshaler_GetZeroValue(t *testing.T) {
	tests := []struct {
		Name      string
		Type      reflect.Type
		Array     bool
		Nullable  bool
		ZeroValue interface{}
	}{
		{
			Name: "str",
			Type: reflect.TypeOf(""),

			ZeroValue: "",
		},
		{
			Name: "str arr",
			Type: reflect.TypeOf(""),

			Array:     true,
			ZeroValue: []string{},
		},
		{
			Name: "str ptr",
			Type: reflect.TypeOf(""),

			Nullable:  true,
			ZeroValue: (*string)(nil),
		},
		{
			Name: "str array ptr",
			Type: reflect.TypeOf(""),

			Array:     true,
			Nullable:  true,
			ZeroValue: (*[]string)(nil),
		},
		{
			Name: "str arr",
			Type: reflect.TypeOf([]string{}),

			ZeroValue: []string{},
		},
		{
			Name: "str arr arr",
			Type: reflect.TypeOf([]string{}),

			Array:     true,
			ZeroValue: [][]string{},
		},
		{
			Name: "str arr ptr",
			Type: reflect.TypeOf([]string{}),

			Nullable:  true,
			ZeroValue: (*[]string)(nil),
		},
		{
			Name: "str arr arr ptr",
			Type: reflect.TypeOf([]string{}),

			Array:     true,
			Nullable:  true,
			ZeroValue: (*[][]string)(nil),
		},
		{
			Name: "testObjType",
			Type: reflect.TypeOf(testObjType{}),

			ZeroValue: testObjType{},
		},
		{
			Name: "testObjType arr",
			Type: reflect.TypeOf(testObjType{}),

			Array:     true,
			ZeroValue: []testObjType{},
		},
		{
			Name: "testObjType ptr",
			Type: reflect.TypeOf(testObjType{}),

			Nullable:  true,
			ZeroValue: (*testObjType)(nil),
		},
		{
			Name: "testObjType array ptr",
			Type: reflect.TypeOf(testObjType{}),

			Array:     true,
			Nullable:  true,
			ZeroValue: (*[]testObjType)(nil),
		},
		{
			Name: "anon struct",
			Type: reflect.TypeOf(struct {
				Prop string
			}{}),

			ZeroValue: struct {
				Prop string
			}{},
		},
		{
			Name: "anon struct arr",
			Type: reflect.TypeOf(struct {
				Prop string
			}{}),

			Array: true,
			ZeroValue: []struct {
				Prop string
			}{},
		},
		{
			Name: "anon struct ptr",
			Type: reflect.TypeOf(struct {
				Prop string
			}{}),

			Nullable: true,
			ZeroValue: (*struct {
				Prop string
			})(nil),
		},
		{
			Name: "anon struct array ptr",
			Type: reflect.TypeOf(struct {
				Prop string
			}{}),

			Array:    true,
			Nullable: true,
			ZeroValue: (*[]struct {
				Prop string
			})(nil),
		},
		{
			Name: "map",
			Type: reflect.TypeOf(map[string]string{}),

			Array:     false,
			Nullable:  false,
			ZeroValue: map[string]string{},
		},
		{
			Name: "map array",
			Type: reflect.TypeOf(map[string]string{}),

			Array:     true,
			Nullable:  false,
			ZeroValue: []map[string]string{},
		},
		{
			Name: "map ptr",
			Type: reflect.TypeOf(map[string]string{}),

			Array:     false,
			Nullable:  true,
			ZeroValue: (*map[string]string)(nil),
		},
		{
			Name: "map array ptr",
			Type: reflect.TypeOf(map[string]string{}),

			Array:     true,
			Nullable:  true,
			ZeroValue: (*[]map[string]string)(nil),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ru := ReflectTypeUnmarshaler{Type: test.Type}
			assert.Equal(t, test.ZeroValue, ru.GetZeroValue(test.Array, test.Nullable))
		})
	}
}

func TestReflectTypeUnmarshaler_UnmarshalToType(t *testing.T) {
	str := "test! äöü"

	tests := []struct {
		Name     string
		Type     reflect.Type
		Array    bool
		Nullable bool
		Data     []byte
		Value    interface{}
	}{
		{
			Name: "str",
			Type: reflect.TypeOf(""),

			Data:  []byte("\"foo\""),
			Value: "foo",
		},
		{
			Name: "str empty",
			Type: reflect.TypeOf(""),

			Data:  []byte("\"\""),
			Value: "",
		},
		{
			Name: "str array",
			Type: reflect.TypeOf(""),

			Array: true,
			Data:  []byte("[\"foo\"]"),
			Value: []string{"foo"},
		},
		{
			Name: "str empty array",
			Type: reflect.TypeOf(""),

			Array: true,
			Data:  []byte("[]"),
			Value: []string{},
		},
		{
			Name: "str ptr null",
			Type: reflect.TypeOf(""),

			Nullable: true,
			Data:     []byte("null"),
			Value:    (*string)(nil),
		},
		{
			Name: "str ptr",
			Type: reflect.TypeOf(""),

			Nullable: true,
			Data:     []byte("\"test! äöü\""),
			Value:    &str,
		},
		{
			Type: reflect.TypeOf(""),

			Name:     "str array ptr null",
			Array:    true,
			Nullable: true,
			Data:     []byte("null"),
			Value:    (*[]string)(nil),
		},
		{
			Name: "str arr ptr",
			Type: reflect.TypeOf(""),

			Array:    true,
			Nullable: true,
			Data:     []byte("[\"abc\",\"def\"]"),
			Value:    &[]string{"abc", "def"},
		},
		{
			Name: "2d string matrix empty",
			Type: reflect.TypeOf(([][]string)(nil)),

			Data:  []byte("[]"),
			Value: [][]string{},
		},
		{
			Name: "2d string matrix",
			Type: reflect.TypeOf(([][]string)(nil)),

			Data:  []byte("[[\"abc\"],[\"abc\",\"def\"]]"),
			Value: [][]string{{"abc"}, {"abc", "def"}},
		},
		{
			Name: "2d string matrix array ptr",
			Type: reflect.TypeOf(([][]string)(nil)),

			Array:    true,
			Nullable: true,
			Data:     []byte("[[[\"abc\"],[\"abc\",\"def\"]]]"),
			Value:    &[][][]string{{{"abc"}, {"abc", "def"}}},
		},
		{
			Name: "map",
			Type: reflect.TypeOf(map[string]string{}),

			Data:  []byte("{\"foo\":\"bar\"}"),
			Value: map[string]string{"foo": "bar"},
		},
		{
			Name: "map array",
			Type: reflect.TypeOf(map[string]string{}),

			Array: true,
			Data:  []byte("[{\"foo\":\"bar\"}]"),
			Value: []map[string]string{{"foo": "bar"}},
		},
		{
			Name: "map nullable array",
			Type: reflect.TypeOf(map[string]string{}),

			Array:    true,
			Nullable: true,
			Data:     []byte("null"),
			Value:    (*[]map[string]string)(nil),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ru := ReflectTypeUnmarshaler{Type: test.Type}
			v, e := ru.UnmarshalToType(test.Data, test.Array, test.Nullable)

			assert.NoError(t, e)
			assert.Equal(t, test.Value, v)
		})
	}

	t.Run("errors", func(t *testing.T) {
		ru := ReflectTypeUnmarshaler{Type: reflect.TypeOf((*[]string)(nil))}

		v, err := ru.UnmarshalToType(nil, false, false)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType(nil, false, true)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType([]byte("null"), false, false)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType([]byte("null"), true, false)
		assert.Nil(t, v)
		assert.Error(t, err)

		ru = ReflectTypeUnmarshaler{Type: reflect.TypeOf((*[]testObjType)(nil))}

		v, err = ru.UnmarshalToType(nil, false, false)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType(nil, false, true)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType([]byte("null"), false, false)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType([]byte("null"), true, false)
		assert.Nil(t, v)
		assert.Error(t, err)

		v, err = ru.UnmarshalToType([]byte("\"test\""), true, false)
		assert.Nil(t, v)
		assert.Error(t, err)
	})
}
