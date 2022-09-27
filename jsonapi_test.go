package jsonapi_test

import (
	"flag"
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
	Bytes  []byte    `json:"bytes" api:"attr"`

	// Relationships
	To1      string   `json:"to-1" api:"rel,mocktype"`
	To1From1 string   `json:"to-1-from-1" api:"rel,mocktype,to-1-from-1"`
	To1FromX string   `json:"to-1-from-x" api:"rel,mocktype,to-x-from-1"`
	ToX      []string `json:"to-x" api:"rel,mocktype"`
	ToXFrom1 []string `json:"to-x-from-1" api:"rel,mocktype,to-1-from-x"`
	ToXFromX []string `json:"to-x-from-x" api:"rel,mocktype,to-x-from-x"`

	meta  Meta
	links map[string]Link
}

func (mt *mocktype) Meta() Meta {
	return mt.meta
}

func (mt *mocktype) SetMeta(m Meta) {
	mt.meta = m
}

func (mt *mocktype) Links() map[string]Link {
	return mt.links
}

func (mt *mocktype) SetLinks(links map[string]Link) {
	mt.links = links
}

type mockTypeImpl struct {
	ID string `json:"id" api:"mockTypeImpl"`

	// Attributes
	Str string `json:"str" api:"attr"`
	Int int    `json:"int" api:"attr"`

	To1 string   `json:"to-1" api:"rel,mockTypeImpl"`
	ToX []string `json:"to-x" api:"rel,mockTypeImpl"`

	meta  Meta
	links map[string]Link
}

func (mt mockTypeImpl) Attrs() map[string]Attr {
	return map[string]Attr{
		"str": {
			Name: "str",
			Type: AttrTypeString,
		},
		"int": {
			Name: "int",
			Type: AttrTypeInt,
		},
	}
}

func (mt mockTypeImpl) Rels() map[string]Rel {
	return map[string]Rel{
		"to-x": {
			FromName: "to-x",
			ToType:   "mockTypeImpl",
			FromType: "mockTypeImpl",
		},
		"to-1": {
			FromName: "to-1",
			ToType:   "mockTypeImpl",
			ToOne:    true,
			FromType: "mockTypeImpl",
		},
	}
}

func (mt mockTypeImpl) GetType() Type {
	return Type{
		Name:  "mockTypeImpl",
		Attrs: mt.Attrs(),
		Rels:  mt.Rels(),
		NewFunc: func() Resource {
			return &mockTypeImpl{}
		},
	}
}

func (mt mockTypeImpl) Get(key string) interface{} {
	switch key {
	case "id":
		return mt.ID
	case "str":
		return mt.Str
	case "int":
		return mt.Int
	// rels
	case "to-x":
		return RelDataMany{
			Res: Identifiers{
				{ID: mt.ToX[1], Meta: Meta{"key1": "value1"}},
				{ID: mt.ToX[0]},
			},
			Links: map[string]Link{
				"example": {HRef: "https://example.org"},
			},
			Meta: Meta{
				"test": "ok",
			},
		}
	case "to-1":
		return RelData{
			Res: Identifier{ID: mt.To1, Meta: map[string]interface{}{
				"k2": "v2",
			}},
			Links: map[string]Link{
				"l1": {HRef: "https://example.org/l1"},
			},
			Meta: Meta{
				"k1": "v1",
			},
		}
	}

	return nil
}

func (mt *mockTypeImpl) Set(key string, val interface{}) {
	switch key {
	case "id":
		mt.ID = val.(string)
	case "str":
		mt.Str = val.(string)
	case "int":
		mt.Int = val.(int)
	case "to-x":
		mt.ToX = val.([]string)
	case "to-1":
		mt.To1 = val.(string)
	}
}

func (mt *mockTypeImpl) Meta() Meta {
	return mt.meta
}

func (mt *mockTypeImpl) SetMeta(meta Meta) {
	mt.meta = meta
}

func (mt *mockTypeImpl) Links() map[string]Link {
	return mt.links
}

func (mt *mockTypeImpl) SetLinks(links map[string]Link) {
	mt.links = links
}
