package jsonapi_test

import (
	"encoding/json"
	"time"

	. "github.com/mark-hartmann/jsonapi"
)

const (
	AttrTypeTestObject = iota + 1
	AttrTypeFloat32Matrix
)

func float32MatrixZeroValue(_ int, array, nullable bool) interface{} {
	switch {
	case array && nullable:
		return (*[][][]float32)(nil)
	case array:
		return [][][]float32{}
	case nullable:
		return (*[][]float32)(nil)
	default:
		return [][]float32{}
	}
}

func float32MatrixUnmarshaler(data []byte, attr Attr) (interface{}, error) {
	if attr.Nullable && string(data) == "null" {
		return float32MatrixZeroValue(attr.Type, attr.Array, attr.Nullable), nil
	}

	var (
		val interface{}
		err error
	)

	if attr.Array {
		var ta [][][]float32
		err = json.Unmarshal(data, &ta)

		if attr.Nullable {
			val = &ta
		} else {
			val = ta
		}
	} else {
		var tObj [][]float32
		err = json.Unmarshal(data, &tObj)

		if attr.Nullable {
			val = &tObj
		} else {
			val = tObj
		}
	}

	if err != nil {
		return nil, err
	}

	return val, nil
}

func testObjectZeroValue(_ int, array, nullable bool) interface{} {
	switch {
	case array && nullable:
		return (*[]testObjType)(nil)
	case array:
		return []testObjType{}
	case nullable:
		return (*testObjType)(nil)
	default:
		return testObjType{}
	}
}

func testObjectUnmarshaler(data []byte, attr Attr) (interface{}, error) {
	if attr.Nullable && string(data) == "null" {
		return testObjectZeroValue(attr.Type, attr.Array, attr.Nullable), nil
	}

	var (
		val interface{}
		err error
	)

	if attr.Array {
		var ta []testObjType
		err = json.Unmarshal(data, &ta)

		if attr.Nullable {
			val = &ta
		} else {
			val = ta
		}
	} else {
		var tObj testObjType
		err = json.Unmarshal(data, &tObj)

		if attr.Nullable {
			val = &tObj
		} else {
			val = tObj
		}
	}

	if err != nil {
		return nil, err
	}

	return val, nil
}

func init() {
	RegisterAttrType(AttrTypeTestObject, "testObject", testObjectZeroValue,
		testObjectUnmarshaler)
	RegisterAttrType(AttrTypeFloat32Matrix, "float32Matrix", float32MatrixZeroValue,
		float32MatrixUnmarshaler)
}

// newMockSchema ...
func newMockSchema() *Schema {
	schema := &Schema{}

	typ := MustBuildType(mockType1{})
	_ = schema.AddType(typ)
	typ = MustBuildType(mockType2{})
	_ = schema.AddType(typ)
	typ = MustBuildType(mockType3{})
	_ = schema.AddType(typ)

	for t, typ := range schema.Types {
		for r, rel := range typ.Rels {
			invType := schema.GetType(rel.ToType)
			rel := schema.Types[t].Rels[r]
			rel.FromOne = invType.Rels[rel.ToName].ToOne
			schema.Types[t].Rels[r] = rel
		}
	}

	errs := schema.Check()
	if len(errs) > 0 {
		panic(errs[0])
	}

	return schema
}

// mockType1 ...
type mockType1 struct {
	ID string `json:"id" api:"mocktypes1"`

	// Attributes
	Str    string    `json:"str" api:"attr"`
	Int    int       `json:"int" api:"attr"`
	Int8   int8      `json:"int8" api:"attr"`
	Int16  int16     `json:"int16" api:"attr"`
	Int32  int32     `json:"int32" api:"attr"`
	Int64  int64     `json:"int64" api:"attr"`
	Uint   uint      `json:"uint" api:"attr"`
	Uint8  uint8     `json:"uint8" api:"attr"`
	Uint16 uint16    `json:"uint16" api:"attr"`
	Uint32 uint32    `json:"uint32" api:"attr"`
	Uint64 uint64    `json:"uint64" api:"attr"`
	Bool   bool      `json:"bool" api:"attr"`
	Time   time.Time `json:"time" api:"attr"`

	// Relationships
	ToOne          string   `json:"to-one" api:"rel,mocktypes2"`
	ToOneFromOne   string   `json:"to-one-from-one" api:"rel,mocktypes2,to-one-from-one"`
	ToOneFromMany  string   `json:"to-one-from-many" api:"rel,mocktypes2,to-many-from-one"`
	ToMany         []string `json:"to-many" api:"rel,mocktypes2"`
	ToManyFromOne  []string `json:"to-many-from-one" api:"rel,mocktypes2,to-one-from-many"`
	ToManyFromMany []string `json:"to-many-from-many" api:"rel,mocktypes2,to-many-from-many"`
}

// mockType2 ...
type mockType2 struct {
	ID string `json:"id" api:"mocktypes2"`

	// Attributes
	StrPtr    *string    `json:"strptr" api:"attr"`
	IntPtr    *int       `json:"intptr" api:"attr"`
	Int8Ptr   *int8      `json:"int8ptr" api:"attr"`
	Int16Ptr  *int16     `json:"int16ptr" api:"attr"`
	Int32Ptr  *int32     `json:"int32ptr" api:"attr"`
	Int64Ptr  *int64     `json:"int64ptr" api:"attr"`
	UintPtr   *uint      `json:"uintptr" api:"attr"`
	Uint8Ptr  *uint8     `json:"uint8ptr" api:"attr"`
	Uint16Ptr *uint16    `json:"uint16ptr" api:"attr"`
	Uint32Ptr *uint32    `json:"uint32ptr" api:"attr"`
	Uint64Ptr *uint64    `json:"uint64ptr" api:"attr"`
	BoolPtr   *bool      `json:"boolptr" api:"attr"`
	TimePtr   *time.Time `json:"timeptr" api:"attr"`

	// Relationships
	ToOneFromOne   string   `json:"to-one-from-one" api:"rel,mocktypes1,to-one-from-one"`
	ToOneFromMany  string   `json:"to-one-from-many" api:"rel,mocktypes1,to-many-from-one"`
	ToManyFromOne  []string `json:"to-many-from-one" api:"rel,mocktypes1,to-one-from-many"`
	ToManyFromMany []string `json:"to-many-from-many" api:"rel,mocktypes1,to-many-from-many"`
}

// mockType3 ...
type mockType3 struct {
	ID string `json:"id" api:"mocktypes3"`

	// Attributes
	Attr1 string `json:"attr1" api:"attr"`
	Attr2 int    `json:"attr2" api:"attr"`

	// Relationships
	Rel1 string   `json:"rel1" api:"rel,mocktypes1"`
	Rel2 []string `json:"rel2" api:"rel,mocktypes1"`
}

type mockType4 struct {
	ID string `json:"id" api:"mocktype4"`

	// Attributes
	StrArr     []string    `json:"strarr" api:"attr"`
	IntArr     []int       `json:"intarr" api:"attr"`
	Int8Arr    []int8      `json:"int8arr" api:"attr"`
	Int16Arr   []int16     `json:"int16arr" api:"attr"`
	Int32Arr   []int32     `json:"int32arr" api:"attr"`
	Int64Arr   []int64     `json:"int64arr" api:"attr"`
	UintArr    []uint      `json:"uintarr" api:"attr"`
	Uint8Arr   []uint8     `json:"uint8arr" api:"attr,bytes"`
	Uint16Arr  []uint16    `json:"uint16arr" api:"attr"`
	Uint32Arr  []uint32    `json:"uint32arr" api:"attr"`
	Uint64Arr  []uint64    `json:"uint64arr" api:"attr"`
	Float32Arr []float32   `json:"float32arr" api:"attr"`
	Float64Arr []float64   `json:"float64arr" api:"attr"`
	BoolArr    []bool      `json:"boolarr" api:"attr"`
	TimeArr    []time.Time `json:"timearr" api:"attr"`
}

type mockType5 struct {
	ID string `json:"id" api:"mocktype5"`

	// Attributes
	StrArrPtr     *[]string    `json:"strarrptr" api:"attr"`
	IntArrPtr     *[]int       `json:"intarrptr" api:"attr"`
	Int8ArrPtr    *[]int8      `json:"int8arrptr" api:"attr"`
	Int16ArrPtr   *[]int16     `json:"int16arrptr" api:"attr"`
	Int32ArrPtr   *[]int32     `json:"int32arrptr" api:"attr"`
	Int64ArrPtr   *[]int64     `json:"int64arrptr" api:"attr"`
	UintArrPtr    *[]uint      `json:"uintarrptr" api:"attr"`
	Uint8ArrPtr   *[]uint8     `json:"uint8arrptr" api:"attr,bytes"`
	Uint16ArrPtr  *[]uint16    `json:"uint16arrptr" api:"attr"`
	Uint32ArrPtr  *[]uint32    `json:"uint32arrptr" api:"attr"`
	Uint64ArrPtr  *[]uint64    `json:"uint64arrptr" api:"attr"`
	Float32ArrPtr *[]float32   `json:"float32arrptr" api:"attr"`
	Float64ArrPtr *[]float64   `json:"float64arrptr" api:"attr"`
	BoolArrPtr    *[]bool      `json:"boolarrptr" api:"attr"`
	TimeArrPtr    *[]time.Time `json:"timearrptr" api:"attr"`
}

type mockType6 struct {
	ID string `json:"ID" api:"mocktype6"`

	Str           string         `json:"str" api:"attr"`
	StrPtr        *string        `json:"strPtr" api:"attr"`
	StrArr        []string       `json:"strArr" api:"attr"`
	StrPtrArr     *[]string      `json:"strPtrArr" api:"attr"`
	Obj           testObjType    `json:"obj" api:"attr,testObject"`
	ObjPtr        *testObjType   `json:"objPtr" api:"attr,testObject"`
	ObjArr        []testObjType  `json:"objArr" api:"attr,testObject"`
	ObjArrPtr     *[]testObjType `json:"objArrPtr" api:"attr,testObject"`
	Float32Matrix [][]float32    `json:"float32Matrix" api:"attr,float32Matrix,no-array"`
}
