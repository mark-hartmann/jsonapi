package jsonapi_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

// TestDocument ...
func TestDocument(t *testing.T) {
	assert := assert.New(t)

	pl1 := Document{}
	assert.Equal(nil, pl1.Data, "empty")
}

func TestInclude(t *testing.T) {
	assert := assert.New(t)

	doc := &Document{}
	typ1 := &Type{Name: "t1"}
	typ2 := &Type{Name: "t2"}

	/*
	 * Main data is a resource
	 */
	doc.Data = newResource(typ1, "id1")

	// Inclusions
	doc.Include(newResource(typ1, "id1"))
	doc.Include(newResource(typ1, "id2"))
	doc.Include(newResource(typ1, "id3"))
	doc.Include(newResource(typ1, "id3"))
	doc.Include(newResource(typ2, "id1"))

	// Check
	ids := []string{}
	for _, res := range doc.Included {
		ids = append(ids, res.GetType().Name+"-"+res.Get("id").(string))
	}

	expect := []string{
		"t1-id2",
		"t1-id3",
		"t2-id1",
	}
	assert.Equal(expect, ids)

	/*
	 * Main data is a collection
	 */
	doc = &Document{}

	// Collection
	col := &SoftCollection{}
	col.SetType(typ1)
	col.Add(newResource(typ1, "id1"))
	col.Add(newResource(typ1, "id2"))
	col.Add(newResource(typ1, "id3"))
	doc.Data = Collection(col)

	// Inclusions
	doc.Include(newResource(typ1, "id1"))
	doc.Include(newResource(typ1, "id2"))
	doc.Include(newResource(typ1, "id3"))
	doc.Include(newResource(typ1, "id4"))
	doc.Include(newResource(typ2, "id1"))
	doc.Include(newResource(typ2, "id1"))
	doc.Include(newResource(typ2, "id2"))

	// Check
	ids = []string{}
	for _, res := range doc.Included {
		ids = append(ids, res.GetType().Name+"-"+res.Get("id").(string))
	}

	expect = []string{
		"t1-id4",
		"t2-id1",
		"t2-id2",
	}
	assert.Equal(expect, ids)
}

func TestMarshalDocumentLinks(t *testing.T) {
	tests := map[string]struct {
		doc *Document
		url *URL
	}{
		"resource with links": {
			doc: &Document{
				Data: &mockTypeImpl{
					ID:  "id1",
					Str: "str",
					Int: 12,
					ToX: []string{
						"id2",
						"id3",
					},
					To1: "id5",
					links: map[string]Link{
						"test": {HRef: "https://example.org/test"},
					},
					meta: map[string]interface{}{
						"foo": "bar",
					},
				},
				RelData: map[string][]string{
					"mockTypeImpl": {"to-x", "to-1"},
				},
			},
			url: &URL{
				Fragments: []string{"fake", "path"},
				Params: &Params{
					Fields: map[string][]string{"mockTypeImpl": {"str", "int", "to-x", "to-1"}},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			// Marshaling
			payload := &bytes.Buffer{}
			err := MarshalDocument(payload, test.doc, test.url)
			assert.NoError(err)

			// Golden file
			filename := strings.ReplaceAll(name, " ", "_") + ".json"
			path := filepath.Join("testdata", "goldenfiles", "marshaling", filename)

			// Retrieve the expected result from file
			expected, _ := ioutil.ReadFile(path)
			assert.NoError(err, name)
			assert.JSONEq(string(expected), payload.String())
		})
	}
}

func TestMarshalDocument(t *testing.T) {
	// TODO Describe how this test suite works
	// Setup
	typ, _ := BuildType(mocktype{})
	typ.NewFunc = func() Resource {
		return Wrap(&mocktype{})
	}
	col := &Resources{}
	col.Add(Wrap(&mocktype{
		ID:       "id1",
		Str:      "str",
		Int:      10,
		Int8:     18,
		Int16:    116,
		Int32:    132,
		Int64:    164,
		Uint:     100,
		Uint8:    108,
		Uint16:   1016,
		Uint32:   1032,
		Uint64:   1064,
		Float32:  math.MaxFloat32,
		Float64:  math.MaxFloat64,
		Bool:     true,
		Time:     getTime(),
		Bytes:    []byte{1, 2, 3},
		To1:      "id2",
		To1From1: "id3",
		To1FromX: "id3",
		ToX:      []string{"id2", "id3"},
		ToXFrom1: []string{"id4"},
		ToXFromX: []string{"id3", "id4"},
	}))
	col.Add(Wrap(&mocktype{
		ID:    "id2",
		Str:   "漢語",
		Int:   -42,
		Time:  time.Time{},
		Bytes: []byte{},
	}))
	col.Add(Wrap(&mocktype{ID: "id3"}))

	r4 := &mocktype{
		ID: "id4",
	}
	r4.SetMeta(map[string]interface{}{
		"key1": "a string",
		"key2": 42,
		"key3": true,
		"key4": getTime(),
	})
	col.Add(Wrap(r4))

	// The collection is ordered by id within the Range function, so this resource
	// will be second in the resource array (0=id1, 1=id1 (this), 2=id2).
	col.Add(Wrap(&mockType1{
		ID:  "id1",
		Str: "str with <html> chars",
	}))

	col.Add(Wrap(&mockType6{
		ID: "test-123",
		Obj: testObjType{
			Prop1: "abc",
			Prop2: "def",
			Prop3: "ghi",
		},
		ObjPtr: &testObjType{
			Prop1: "jkl",
			Prop2: "mno",
			Prop3: "pqr",
		},
		Float32Matrix: [][]float32{
			{1.0, 0.5, 0.25, 0.175},
			{0.175, 0.25, 0.5, 1.0},
		},
	}))

	var (
		strarr   = []string{"foo", "bar", "baz"}
		int8arr  = []int8{-100, -50, 0, 50, 100}
		int32arr = []int32{-10000000, 123456, -45, 333333333}
		bytearr  = []byte("hello world")
		boolarr  = []bool{true, false, true, true}
	)

	cres1 := Wrap(mockType4{
		ID:         "id1",
		StrArr:     strarr,
		Int8Arr:    int8arr,
		Int32Arr:   int32arr,
		Uint8Arr:   bytearr,
		BoolArr:    boolarr,
		Float32Arr: []float32{math.MaxFloat32},
		Float64Arr: []float64{math.MaxFloat64},
	})

	cres2 := Wrap(mockType5{
		ID:            "id1",
		StrArrPtr:     &strarr,
		Int8ArrPtr:    &int8arr,
		BoolArrPtr:    &boolarr,
		Float32ArrPtr: &[]float32{3.4028235e+38},
	})

	// uint8
	cres3 := &SoftResource{Type: &Type{
		Name: "bytestest",
		Attrs: map[string]Attr{
			"uint8arr": {
				Name:  "uint8arr",
				Type:  AttrTypeUint8,
				Array: true,
			},
			"uint8arrptr": {
				Name:     "uint8arrptr",
				Type:     AttrTypeUint8,
				Array:    true,
				Nullable: true,
			},
			"uint8arrempty": {
				Name:  "uint8arrempty",
				Type:  AttrTypeUint8,
				Array: true,
			},
			"uint8arrptrnull": {
				Name:     "uint8arrptrnull",
				Type:     AttrTypeUint8,
				Array:    true,
				Nullable: true,
			},
			"bytes": {
				Name: "bytes",
				Type: AttrTypeBytes,
			},
			"bytesptr": {
				Name:     "bytesptr",
				Type:     AttrTypeBytes,
				Nullable: true,
			},
			"nullbytes": {
				Name: "nullbytes",
				Type: AttrTypeBytes,
			},
			"nullbytesptr": {
				Name:     "nullbytesptr",
				Type:     AttrTypeBytes,
				Nullable: true,
			},
		},
	}}

	arr := []uint8{1, 2, 4, 8, 16, 32}

	cres3.SetID("id1")
	cres3.Set("uint8arr", arr)
	cres3.Set("uint8arrptr", &arr)
	cres3.Set("uint8arrempty", nil)
	cres3.Set("uint8arrptrnull", nil)
	cres3.Set("bytes", arr)
	cres3.Set("bytesptr", &arr)

	cres4 := &SoftResource{Type: &Type{
		Name: "objtest",
		Attrs: map[string]Attr{
			"obj": {
				Name: "obj",
				Type: AttrTypeTestObject,
			},
			"objarr": {
				Name:  "objarr",
				Type:  AttrTypeTestObject,
				Array: true,
			},
			"objptr": {
				Name:     "objptr",
				Type:     AttrTypeTestObject,
				Nullable: true,
			},
			"float32MatrixArr": {
				Name:  "float32MatrixArr",
				Type:  AttrTypeFloat32Matrix,
				Array: true,
			},
		},
	}}
	cres4.SetID("id1")
	cres4.Set("obj", testObjType{
		Prop1: "foo",
		Prop2: "bar",
		Prop3: "baz",
	})
	cres4.Set("objarr", []testObjType{
		{
			Prop1: "a",
			Prop2: "b",
			Prop3: "c",
		},
		{
			Prop1: "d",
			Prop2: "e",
			Prop3: "f",
		},
	})
	cres4.Set("objptr", nil)
	cres4.Set("float32MatrixArr", [][][]float32{
		{
			{0.1, 0.2, 0.3, 0.4, 0.5},
		},
		{
			{0.6, 0.7, 0.8, 0.9, 1.0},
		},
	})

	// Test struct
	tests := []struct {
		name   string
		doc    *Document
		fields map[string][]string
	}{
		{
			name: "empty data",
			doc: &Document{
				PrePath: "https://example.org",
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "empty data with links",
			doc: &Document{
				PrePath: "https://example.org",
				Links: map[string]Link{
					"foo": {HRef: "https://example.org/bar", Meta: map[string]interface{}{
						"last_accessed": "5 minutes ago",
					}},
					"bar": {HRef: "https://example.org/baz"},
				},
			},
		}, {
			name: "empty collection",
			doc: &Document{
				Data: &Resources{},
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "resource",
			doc: &Document{
				Data: col.At(0),
				RelData: map[string][]string{
					"mocktype": {"to-1", "to-x-from-1"},
				},
			},
			fields: map[string][]string{
				"mocktype": {
					"str",
					"uint64",
					"bool",
					"int",
					"time",
					"bytes",
					"float32",
					"float64",
					"to-1",
					"to-x-from-1",
				},
			},
		}, {
			name: "resource array attributes",
			doc: &Document{
				Data: cres1,
			},
			fields: map[string][]string{
				"mocktype4": {
					"strarr",
					"int8arr",
					"int32arr",
					"uint8arr",
					"boolarr",
					"int16arr",
					"float32arr",
					"float64arr",
				},
			},
		}, {
			name: "resource nullable array attributes",
			doc: &Document{
				Data: cres2,
			},
			fields: map[string][]string{
				"mocktype5": {
					"strarrptr",
					"int8arrptr",
					"int32arrptr",
					"uint8arrptr",
					"boolarrptr",
					"int16arrptr",
					"float32arrptr",
					"float64arrptr",
				},
			},
		}, {
			name: "resource bytes",
			doc: &Document{
				Data: cres3,
			},
			fields: map[string][]string{
				"bytestest": {
					"uint8arr",
					"uint8arrptr",
					"bytes",
					"bytesptr",
					"nullbytes",
					"nullbytesptr",
					"uint8arrempty",
					"uint8arrptrnull",
				},
			},
		}, {
			name: "resource object property",
			doc: &Document{
				Data: cres4,
			},
			fields: map[string][]string{
				"objtest": {"obj", "objarr", "objptr", "objptrarr", "float32MatrixArr"},
			},
		}, {
			name: "collection",
			doc: &Document{
				Data: col,
				RelData: map[string][]string{
					"mocktype": {"to-1", "to-x-from-1"},
				},
				PrePath: "https://example.org",
			},
			fields: map[string][]string{
				"mocktype":   {"str", "uint64", "bool", "int", "time", "to-1", "to-x-from-1"},
				"mocktypes1": {"str"},
				"mocktype6":  {"obj", "objPtr", "objArr", "float32Matrix"},
			},
		}, {
			name: "collection with links",
			doc: &Document{
				Data: col,
				RelData: map[string][]string{
					"mocktype": {"to-1", "to-x-from-1"},
				},
				PrePath: "https://example.org",
				Links: map[string]Link{
					"foo": {
						HRef: "https://example.org/bar",
						Meta: map[string]interface{}{
							"foo": "bar",
							"bar": "baz",
						},
					},
				},
			},
			fields: map[string][]string{
				"mocktype": {"str", "uint64", "bool", "int", "time", "to-1", "to-x-from-1"},
			},
		}, {
			name: "meta",
			doc: &Document{
				Data: nil,
				Meta: map[string]interface{}{
					"f1": "漢語",
					"f2": 42,
					"f3": true,
				},
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "collection with inclusions",
			doc: &Document{
				Data: Wrap(&mocktype{
					ID: "id1",
				}),
				RelData: map[string][]string{
					"mocktype": {"to-1", "to-x-from-1"},
				},
				Included: []Resource{
					Wrap(&mocktype{
						ID: "id2",
					}),
					Wrap(&mocktype{
						ID: "id3",
					}),
					Wrap(&mocktype{
						ID: "id4",
					}),
				},
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "identifier",
			doc: &Document{
				Data: Identifier{
					ID:   "id1",
					Type: "mocktype",
				},
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "identifiers",
			doc: &Document{
				Data: Identifiers{
					{
						ID:   "id1",
						Type: "mocktype",
					}, {
						ID:   "id2",
						Type: "mocktype",
					}, {
						ID:   "id3",
						Type: "mocktype",
					},
				},
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "error",
			doc: &Document{
				Errors: func() []Error {
					err := NewErrBadRequest("Bad Request", "This request is bad.")
					err.ID = "00000000-0000-0000-0000-000000000000"
					return []Error{err}
				}(),
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		}, {
			name: "errors",
			doc: &Document{
				Errors: func() []Error {
					err1 := NewErrBadRequest("Bad Request", "This request is bad.")
					err1.ID = "00000000-0000-0000-0000-000000000000"
					err2 := NewErrBadRequest("Bad Request", "This request is really bad.")
					err2.ID = "00000000-0000-0000-0000-000000000000"
					return []Error{err1, err2}
				}(),
			},
			fields: map[string][]string{
				"mocktype": nil, // To achieve the same result as before with []string
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// URL
			url := &URL{
				Fragments: []string{"fake", "path"},
				Params: &Params{
					Fields: test.fields,
				},
			}
			if _, ok := test.doc.Data.(Collection); ok {
				url.IsCol = true
			}

			// Marshaling
			payload := &bytes.Buffer{}
			err := MarshalDocument(payload, test.doc, url)
			assert.NoError(t, err)

			// Golden file
			filename := strings.ReplaceAll(test.name, " ", "_") + ".json"
			path := filepath.Join("testdata", "goldenfiles", "marshaling", filename)
			if !*update {
				// Retrieve the expected result from file
				expected, _ := ioutil.ReadFile(path)
				assert.NoError(t, err, test.name)
				assert.JSONEq(t, string(expected), payload.String())
			} else {
				dst := &bytes.Buffer{}
				err = json.Indent(dst, payload.Bytes(), "", "\t")
				assert.NoError(t, err)
				// TODO Figure out whether 0600 is okay or not.
				err = ioutil.WriteFile(path, dst.Bytes(), 0600)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalInvalidDocuments(t *testing.T) {
	// TODO Describe how this test suite works
	// Setup
	typ, _ := BuildType(mocktype{})
	typ.NewFunc = func() Resource {
		return Wrap(&mocktype{})
	}
	col := &Resources{}
	col.Add(Wrap(&mocktype{
		ID:       "id1",
		Str:      "str",
		Int:      10,
		Int8:     18,
		Int16:    116,
		Int32:    132,
		Int64:    164,
		Uint:     100,
		Uint8:    108,
		Uint16:   1016,
		Uint32:   1032,
		Uint64:   1064,
		Bool:     true,
		Time:     getTime(),
		To1:      "id2",
		To1From1: "id3",
		To1FromX: "id3",
		ToX:      []string{"id2", "id3"},
		ToXFrom1: []string{"id4"},
		ToXFromX: []string{"id3", "id4"},
	}))
	col.Add(Wrap(&mocktype{
		ID:   "id2",
		Str:  "漢語",
		Int:  -42,
		Time: time.Time{},
	}))
	col.Add(Wrap(&mocktype{ID: "id3"}))

	// Test struct
	tests := []struct {
		name   string
		doc    *Document
		fields []string
		err    string
	}{
		{
			name: "invalid data",
			doc: &Document{
				Data: "just a string",
			},
			err: "data contains an unknown type",
		},
	}

	for i := range tests {
		i := i
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// URL
			url := &URL{
				Fragments: []string{"fake", "path"},
				Params: &Params{
					Fields: map[string][]string{"mocktype": test.fields},
				},
			}
			if _, ok := test.doc.Data.(Collection); ok {
				url.IsCol = true
			}

			// Marshaling
			w := &bytes.Buffer{}
			err := MarshalDocument(w, test.doc, url)
			assert.EqualError(err, test.err)
		})
	}
}

func TestUnmarshalDocument(t *testing.T) {
	// Setup
	typ, _ := BuildType(mocktype{})
	typ.NewFunc = func() Resource {
		return Wrap(&mocktype{})
	}

	typ4, _ := BuildType(mockType4{})
	typ4.NewFunc = func() Resource {
		return Wrap(&mockType4{})
	}

	typ5, _ := BuildType(mockType5{})
	typ5.NewFunc = func() Resource {
		return Wrap(&mockType5{})
	}

	schema := &Schema{Types: []Type{typ, typ4, typ5}}
	col := Resources{}
	col.Add(Wrap(&mocktype{
		ID:       "id1",
		Str:      "str",
		Int:      10,
		Int8:     18,
		Int16:    116,
		Int32:    132,
		Int64:    164,
		Uint:     100,
		Uint8:    108,
		Uint16:   1016,
		Uint32:   1032,
		Uint64:   1064,
		Float32:  math.MaxFloat32,
		Float64:  math.MaxFloat64,
		Bool:     true,
		Time:     getTime(),
		Bytes:    []byte{1, 2, 3},
		To1:      "id2",
		To1From1: "id3",
		To1FromX: "id3",
		ToX:      []string{"id2", "id3"},
		ToXFrom1: []string{"id4"},
		ToXFromX: []string{"id3", "id4"},
	}))
	col.Add(Wrap(&mocktype{ID: "id2"}))
	col.Add(Wrap(&mocktype{ID: "id3"}))
	col.Add(Wrap(&mockType4{
		ID:      "id1",
		BoolArr: []bool{true, false},
		StrArr:  []string{"a", "b", "c"},
	}))
	col.Add(Wrap(&mockType5{
		ID:          "id123",
		BoolArrPtr:  &[]bool{true, false},
		StrArrPtr:   &[]string{"a", "b", "c"},
		Uint8ArrPtr: &[]byte{1, 2, 4, 8, 16},
	}))

	uint8arrRes := &SoftResource{}
	uint8arrRes.SetType(&Type{
		Name: "uint8arrtest",
		Attrs: map[string]Attr{
			"uint8arr": {Name: "uint8arr", Type: AttrTypeUint8, Array: true},
			"bytes":    {Name: "bytes", Type: AttrTypeBytes},
		},
	})

	_ = schema.AddType(*uint8arrRes.Type)

	uint8arrRes.SetID("id1")
	uint8arrRes.Set("uint8arr", []uint8{0, 1, 2, 4, 8, 16, 32, 64, 128, 255})
	uint8arrRes.Set("bytes", []uint8{0, 1, 2, 4, 8, 16, 32, 64, 128, 255})

	col.Add(uint8arrRes)

	objRes := &SoftResource{}
	objRes.SetType(&Type{
		Name: "objtest",
		Attrs: map[string]Attr{
			"obj": {
				Name: "obj",
				Type: AttrTypeTestObject,
			},
			"objarr": {
				Name:  "objarr",
				Type:  AttrTypeTestObject,
				Array: true,
			},
			"objptr": {
				Name:     "objptr",
				Type:     AttrTypeTestObject,
				Nullable: true,
			},
			"objptrarr": {
				Name:     "objptrarr",
				Type:     AttrTypeTestObject,
				Array:    true,
				Nullable: true,
			},
		},
	})
	objRes.SetID("id1")
	objRes.Set("obj", testObjType{
		Prop1: "foo",
		Prop2: "bar",
		Prop3: "baz",
	})
	objRes.Set("objarr", []testObjType{
		{
			Prop1: "a",
			Prop2: "b",
			Prop3: "c",
		},
		{
			Prop1: "c",
			Prop2: "d",
			Prop3: "e",
		},
	})
	objRes.Set("objptr", nil)
	objRes.Set("objptrarr", []testObjType{
		{
			Prop1: "1",
			Prop2: "2",
			Prop3: "3",
		},
	})

	_ = schema.AddType(*objRes.Type)

	r4 := &mocktype{
		ID: "id4",
	}
	r4.SetMeta(map[string]interface{}{
		"key1": "a string",
		"key2": 42,
		"key3": true,
		"key4": getTime(),
	})
	col.Add(Wrap(r4))

	r6 := Wrap(&mockType6{
		ID: "test-123",
		Obj: testObjType{
			Prop1: "abc",
			Prop2: "def",
			Prop3: "ghi",
		},
		ObjPtr: &testObjType{
			Prop1: "jkl",
			Prop2: "mno",
			Prop3: "pqr",
		},
	})
	col.Add(r6)
	typ6 := r6.GetType()

	_ = schema.AddType(typ6)

	// Tests
	t.Run("resource with inclusions", func(t *testing.T) {
		assert := assert.New(t)

		url, _ := NewURLFromRaw(schema, "/mocktype/id1?fields[mocktype]="+
			strings.Join(typ.Fields(), ","))

		doc := &Document{
			Data: col.At(0),
			RelData: map[string][]string{
				"mocktype": typ.Fields(),
			},
			Included: []Resource{
				col.At(1),
				col.At(2),
			},
		}

		payload := &bytes.Buffer{}
		err := MarshalDocument(payload, doc, url)
		assert.NoError(err)

		doc2, err := UnmarshalDocument(payload, schema)
		assert.NoError(err)
		assert.True(Equal(doc.Data.(Resource), doc2.Data.(Resource)))
		// TODO Make all the necessary assertions.
	})

	t.Run("resource with object property", func(t *testing.T) {
		url, _ := NewURLFromRaw(schema, "/objtest/id1?fields[objtest]=obj,objarr,"+
			"objptr,objptrarr")
		doc := &Document{Data: objRes}

		payload := &bytes.Buffer{}
		err := MarshalDocument(payload, doc, url)
		assert.NoError(t, err)

		doc2, err := UnmarshalDocument(payload, schema)
		assert.NoError(t, err)
		assert.True(t, Equal(doc.Data.(Resource), doc2.Data.(Resource)))
	})

	t.Run("collection with inclusions", func(t *testing.T) {
		assert := assert.New(t)

		url, _ := NewURLFromRaw(schema, "/mocktype/id1?fields[mocktype]="+
			strings.Join(typ.Fields(), ","))
		url.Params.Fields["mocktype4"] = typ4.Fields()
		url.Params.Fields["mocktype5"] = typ5.Fields()
		url.Params.Fields["mocktype6"] = typ6.Fields()
		url.Params.Fields["uint8arrtest"] = uint8arrRes.Type.Fields()

		doc := &Document{
			Data: &col,
			RelData: map[string][]string{
				"mocktype": typ.Fields(),
			},
		}

		payload := &bytes.Buffer{}
		err := MarshalDocument(payload, doc, url)
		assert.NoError(err)

		doc2, err := UnmarshalDocument(payload, schema)
		assert.NoError(err)
		assert.IsType(&col, doc.Data)
		assert.IsType(&col, doc2.Data)
		if col, ok := doc.Data.(Collection); ok {
			if col2, ok := doc2.Data.(Collection); ok {
				assert.Equal(col.Len(), col2.Len())
				for j := 0; j < col.Len(); j++ {
					assert.True(Equal(col.At(j), col2.At(j)))
				}
			}
		}

		col2, ok := doc.Data.(Collection)
		assert.True(ok)

		// A few assertions to make sure some edge cases work.
		arrRes := col2.At(5)
		assert.Equal("uint8arrtest", arrRes.GetType().Name)
		assert.Equal([]byte{0, 1, 2, 4, 8, 16, 32, 64, 128, 255}, arrRes.Get("uint8arr"))
		assert.Equal(arrRes.Get("uint8arr"), arrRes.Get("bytes"))

		// TODO Make all the necessary assertions.
	})

	t.Run("errors (Unmarshal)", func(t *testing.T) {
		assert := assert.New(t)

		url, _ := NewURLFromRaw(schema, "/mocktype/id1/relationships/to-x")

		doc := &Document{
			Errors: func() []Error {
				err := NewErrBadRequest("Bad Request", "This request is bad.")
				err.ID = "00000000-0000-0000-0000-000000000000"
				return []Error{err}
			}(),
		}

		payload := &bytes.Buffer{}
		err := MarshalDocument(payload, doc, url)
		assert.NoError(err)

		doc2, err := UnmarshalDocument(payload, schema)
		assert.NoError(err)
		assert.Equal(doc.Data, doc2.Data)
	})
}

func TestUnmarshalDocument_Invalid(t *testing.T) {
	schema := newMockSchema()
	_ = schema.AddType(Type{
		Name: "objtest",
		Attrs: map[string]Attr{
			"obj": {Name: "obj", Type: AttrTypeTestObject},
		},
	})

	t.Run("invalid payload", func(t *testing.T) {
		payload := `invalid payload`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, "invalid character 'i' looking for beginning of value")
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})

	t.Run("invalid data member type", func(t *testing.T) {
		payload := `{"data":"invaliddata"}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, "jsonapi: invalid member data type")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data", src)
	})

	t.Run("missing primary data member", func(t *testing.T) {
		payload := `{"datum":[]}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, "jsonapi: missing primary member")
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})

	t.Run("data and error primary data members", func(t *testing.T) {
		payload := `{"data":[], "errors":[]}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, `jsonapi: "data" and "errors" must not coexist`)
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})

	t.Run("included without data", func(t *testing.T) {
		payload := `{"meta": {}, "included":[{"id":"some-id"}]}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})

	t.Run("invalid id data type", func(t *testing.T) {
		payload := `{"data":{"id":true}}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, "jsonapi: failed to unmarshal resource: json: cannot "+
			"unmarshal bool into Go struct field resourceSkeleton.id of type string")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data", src)
	})

	t.Run("invalid id data type collection", func(t *testing.T) {
		payload := `{"data":[{"id":true}]}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal collection: "+
			"jsonapi: failed to unmarshal resource at 0: "+
			"json: cannot unmarshal bool into Go struct field resourceSkeleton.id of type string")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data/0", src)
	})

	t.Run("data is null and attributes wrong type", func(t *testing.T) {
		payload := `{"data":null,"included":[{"attributes":true}]}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal included resource at 0: "+
			"json: cannot unmarshal bool into Go struct field resourceSkeleton.attributes of "+
			"type map[string]json.RawMessage")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/included/0", src)
	})

	t.Run("unknown attribute for resource type", func(t *testing.T) {
		payload := `{"data":{"id":"1","type":"mocktypes1","attributes":{"nonexistent":1}}}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal resource: "+
			`jsonapi: field "nonexistent" does not exist in resource type "mocktypes1"`)

		var unknownFieldErr *UnknownFieldError
		assert.ErrorAs(t, err, &unknownFieldErr)
		assert.Equal(t, "mocktypes1", unknownFieldErr.Type)
		assert.Equal(t, "nonexistent", unknownFieldErr.Field)
		assert.True(t, unknownFieldErr.IsAttr())
		assert.False(t, unknownFieldErr.InPath())
		assert.Equal(t, "", unknownFieldErr.RelPath())

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data/attributes", src)
	})

	t.Run("invalid attribute type", func(t *testing.T) {
		payload := `{"data":{"id":"1","type":"mocktypes1","attributes":{"int8":"abc"}}}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal resource: "+
			`jsonapi: invalid value "\"abc\"" for field "int8": `+
			`strconv.Atoi: parsing "\"abc\"": invalid syntax`)

		var InvalidFieldValueErr *InvalidFieldValueError
		assert.ErrorAs(t, err, &InvalidFieldValueErr)
		assert.Equal(t, "mocktypes1", InvalidFieldValueErr.Type)
		assert.Equal(t, "int8", InvalidFieldValueErr.Field)
		assert.Equal(t, "int8", InvalidFieldValueErr.FieldType)
		assert.Equal(t, `"abc"`, InvalidFieldValueErr.Value)
		assert.True(t, InvalidFieldValueErr.IsAttr())

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data/attributes/int8", src)
	})

	t.Run("invalid custom attribute type", func(t *testing.T) {
		payload := `
{
	"data": {
		"id": "1",
		"type": "objtest",
		"attributes": {
			"obj": 123
		}
	}
}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal resource: "+
			`jsonapi: invalid value "123" for field "obj": `+
			`json: cannot unmarshal number into Go value of type jsonapi_test.testObjType`)

		var InvalidFieldValueErr *InvalidFieldValueError
		assert.ErrorAs(t, err, &InvalidFieldValueErr)
		assert.Equal(t, "objtest", InvalidFieldValueErr.Type)
		assert.Equal(t, "obj", InvalidFieldValueErr.Field)
		assert.Equal(t, "testObject", InvalidFieldValueErr.FieldType)
		assert.Equal(t, `123`, InvalidFieldValueErr.Value)
		assert.True(t, InvalidFieldValueErr.IsAttr())

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data/attributes/obj", src)
	})

	t.Run("invalid relationship data type", func(t *testing.T) {
		payload := `
{
	"data": {
	"id": "1",
		"type": "mocktypes1",
		"relationships": {
			"to-many-from-one": {
				"data": "wrong"
			}
		}
	}
}`
		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal resource: "+
			"json: cannot unmarshal string into Go value of type jsonapi.Identifiers")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data/relationships/to-many-from-one", src)
	})

	t.Run("unknown relationship", func(t *testing.T) {
		payload := `
{
	"data": {
		"id": "1",
		"type": "mocktypes1",
		"relationships": {
			"wrong": {
				"data": {"id": "foo", "type": "some-type"}
			}
		}
	}
}`
		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal resource: "+
			`jsonapi: field "wrong" does not exist in resource type "mocktypes1"`)

		var unknownFieldErr *UnknownFieldError
		assert.ErrorAs(t, err, &unknownFieldErr)
		assert.Equal(t, "mocktypes1", unknownFieldErr.Type)
		assert.Equal(t, "wrong", unknownFieldErr.Field)
		assert.False(t, unknownFieldErr.IsAttr())
		assert.False(t, unknownFieldErr.InPath())
		assert.Equal(t, "", unknownFieldErr.RelPath())

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/data/relationships", src)
	})

	t.Run("invalid included identifier", func(t *testing.T) {
		payload := `{"data":null,"included":[{"id":true}]}`

		_, err := UnmarshalDocument(strings.NewReader(payload), schema)
		assert.EqualError(t, err, ""+
			"jsonapi: failed to unmarshal included resource at 0: "+
			"json: cannot unmarshal bool into Go struct field resourceSkeleton.id of type string")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var srcErr srcError
		assert.ErrorAs(t, err, &srcErr)

		src, isPtr := srcErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/included/0", src)
	})
}

func newResource(typ *Type, id string) Resource {
	res := &SoftResource{}
	res.SetType(typ)
	res.SetID(id)

	return res
}
