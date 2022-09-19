package jsonapi_test

import (
	"encoding/json"
	"flag"
	"strings"
	"time"

	. "github.com/mfcochauxlaberge/jsonapi"
)

var update = flag.Bool("update-golden-files", false, "update the golden files")

func getTime() time.Time {
	now, _ := time.Parse(time.RFC3339Nano, "2013-06-24T22:03:34.8276Z")
	return now
}

var _ MetaHolder = (*mocktype)(nil)

// mocktype is a fake struct that defines a JSON:API type for test purposes.
type mocktype struct {
	ID string `json:"id" api:"mocktype"`

	// Attributes
	Str     string    `json:"str" api:"attr"`
	Int     int       `json:"int" api:"attr"`
	Int8    int8      `json:"int8" api:"attr"`
	Int16   int16     `json:"int16" api:"attr"`
	Int32   int32     `json:"int32" api:"attr"`
	Int64   int64     `json:"int64" api:"attr"`
	Uint    uint      `json:"uint" api:"attr"`
	Uint8   uint8     `json:"uint8" api:"attr"`
	Uint16  uint16    `json:"uint16" api:"attr"`
	Uint32  uint32    `json:"uint32" api:"attr"`
	Uint64  uint64    `json:"uint64" api:"attr"`
	Float32 float32   `json:"float32" api:"attr"`
	Float64 float64   `json:"float64" api:"attr"`
	Bool    bool      `json:"bool" api:"attr"`
	Time    time.Time `json:"time" api:"attr"`
	Bytes   []byte    `json:"bytes" api:"attr" bytes:"true"`

	// Relationships
	To1      string   `json:"to-1" api:"rel,mocktype"`
	To1From1 string   `json:"to-1-from-1" api:"rel,mocktype,to-1-from-1"`
	To1FromX string   `json:"to-1-from-x" api:"rel,mocktype,to-x-from-1"`
	ToX      []string `json:"to-x" api:"rel,mocktype"`
	ToXFrom1 []string `json:"to-x-from-1" api:"rel,mocktype,to-1-from-x"`
	ToXFromX []string `json:"to-x-from-x" api:"rel,mocktype,to-x-from-x"`

	meta Meta
}

func (mt *mocktype) Meta() Meta {
	return mt.meta
}

func (mt *mocktype) SetMeta(m Meta) {
	mt.meta = m
}

// testObjType is a simple struct used as an attribute type.
type testObjType struct {
	Prop1 string `json:"prop1"`
	Prop2 string `json:"prop2"`
	Prop3 string `json:"prop3"`
}

func (t testObjType) PublicTypeName() string {
	return "test-object#123"
}

func (t testObjType) GetZeroValue(array, nullable bool) interface{} {
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

func (t testObjType) UnmarshalToType(data []byte, array, nullable bool) (interface{}, error) {
	if nullable && string(data) == "null" {
		return t.GetZeroValue(array, nullable), nil
	}

	var (
		val interface{}
		err error
	)

	if array {
		var ta []testObjType
		err = json.Unmarshal(data, &ta)

		if nullable {
			val = &ta
		} else {
			val = ta
		}
	} else {
		var tObj testObjType
		err = json.Unmarshal(data, &tObj)

		if nullable {
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

// stringTypeUnmarshalerRot13 shows that even wacky use-cases can be implemented using
// a TypeUnmarshaler. If this TypeUnmarshaler is used with AttrTypeString, the payload is
// rot-13 encoded.
type stringTypeUnmarshalerRot13 struct {
}

func (c stringTypeUnmarshalerRot13) GetZeroValue(array, nullable bool) interface{} {
	switch {
	case array && nullable:
		return (*[]string)(nil)
	case array:
		return []string{}
	case nullable:
		return (*string)(nil)
	default:
		return ""
	}
}

func (c stringTypeUnmarshalerRot13) UnmarshalToType(data []byte, array, nullable bool) (interface{}, error) {
	var v interface{}
	var err error

	rot13 := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			if r >= 'm' {
				return r - 13
			} else {
				return r + 13
			}
		} else if r >= 'A' && r <= 'Z' {
			if r >= 'M' {
				return r - 13
			} else {
				return r + 13
			}
		}
		return r
	}

	if array {
		var sa []string
		err = json.Unmarshal(data, &sa)

		for i, s := range sa {
			sa[i] = strings.Map(rot13, s)
		}

		if nullable {
			v = &sa
		} else {
			v = sa
		}
	} else if string(data) != "null" {
		var s string
		err = json.Unmarshal(data, &s)
		s = strings.Map(rot13, s)

		if nullable {
			v = &s
		} else {
			v = s
		}
	}

	return v, err
}

