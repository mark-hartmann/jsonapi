package jsonapi_test

import (
	"testing"

	. "github.com/mark-hartmann/jsonapi"

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

func TestParseSortRule(t *testing.T) {
	schema := newMockSchema()

	testData := []struct {
		raw string
		typ Type
		res SortRule
		err bool
	}{
		{
			raw: "",
			typ: schema.GetType("mocktypes1"),
			err: true,
		},
		{
			raw: "-",
			typ: schema.GetType("mocktypes1"),
			err: true,
		},
		{
			raw: "uint16",
			typ: schema.GetType("mocktypes1"),
			res: SortRule{
				Name: "uint16",
			},
		},
		{
			raw: "-uint16",
			typ: schema.GetType("mocktypes1"),
			res: SortRule{
				Name: "uint16",
				Desc: true,
			},
		},
		{
			raw: "to-one-from-one.uint16",
			typ: schema.GetType("mocktypes1"),
			res: SortRule{
				Name: "uint16",
			},
			err: true,
		},
		{
			raw: "-to-one-from-one.int16ptr",
			typ: schema.GetType("mocktypes1"),
			res: SortRule{
				Path: []Rel{
					schema.GetType("mocktypes1").Rels["to-one-from-one"],
				},
				Name: "int16ptr",
				Desc: true,
			},
		},
		{
			raw: "-to-one-from-one.to-one-from-many.int8",
			typ: schema.GetType("mocktypes1"),
			res: SortRule{
				Name: "int8",
				Desc: true,
			},
		},
		{
			raw: "-to-one-from-one.unknown-relationship.some-prop",
			typ: schema.GetType("mocktypes1"),
			err: true,
		},
	}

	for _, test := range testData {
		rule, err := ParseSortRule(schema, test.typ, test.raw)

		if test.err {
			assert.Error(t, err, test.raw)
		} else {
			assert.NoError(t, err, test.raw)
			assert.Equal(t, test.res, rule, test.raw)
		}
	}

	rule := "-to-one-from-one.unknown-relationship.some-prop"
	_, err := ParseSortRule(schema, schema.GetType("mocktypes1"), rule)

	var ufErr *UnknownFieldError

	assert.ErrorAs(t, err, &ufErr)
	assert.Equal(t, "mocktypes2", ufErr.Type)
	assert.Equal(t, "unknown-relationship", ufErr.Field)
	assert.False(t, ufErr.IsAttr())
	assert.False(t, ufErr.InPath())
	assert.Equal(t, rule[1:], ufErr.RelPath())

	rule = "to-one-from-one.uint16"
	_, err = ParseSortRule(schema, schema.GetType("mocktypes1"), rule)

	assert.ErrorAs(t, err, &ufErr)
	assert.Equal(t, "mocktypes2", ufErr.Type)
	assert.Equal(t, "uint16", ufErr.Field)
	assert.True(t, ufErr.IsAttr())
	assert.False(t, ufErr.InPath())
	assert.Equal(t, rule, ufErr.RelPath())

	rule = "-to-many.int32ptr"
	_, err = ParseSortRule(schema, schema.GetType("mocktypes1"), rule)

	var ifErr *InvalidFieldError

	assert.ErrorAs(t, err, &ifErr)
	assert.Equal(t, "mocktypes1", ifErr.Type)
	assert.Equal(t, "to-many", ifErr.Field)
	assert.False(t, ifErr.IsAttr())
	assert.True(t, ifErr.IsInvalidRelType())
	assert.Equal(t, rule[1:], ifErr.RelPath())
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
