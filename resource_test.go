package jsonapi_test

import (
	"testing"
	"time"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalPartialResource(t *testing.T) {
	// Setup
	typ, _ := BuildType(mocktype{})
	typ.NewFunc = func() Resource {
		return Wrap(&mocktype{})
	}
	typ4, err := BuildType(mockType4{})
	assert.NoError(t, err)

	schema := &Schema{Types: []Type{typ, typ4}}

	// Tests
	t.Run("partial resource", func(t *testing.T) {
		assert := assert.New(t)

		payload := `{
			"id": "abc123",
			"type": "mocktype",
			"attributes": {
				"str": "abc",
				"float64": 123.456789
			},
			"relationships": {
				"to-1": {
					"data": {
						"type": "mocktype",
						"data": "def"
					}
				},
				"to-x": {
					"data": [
						{
							"type": "mocktype",
							"data": "ghi"
						},
						{
							"type": "mocktype",
							"data": "jkl"
						}
					]
				}
			}
		}`

		res, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.NoError(err)

		assert.Equal("abc123", res.GetID())
		assert.Equal("mocktype", res.GetType().Name)
		assert.Len(res.Attrs(), 2)
		assert.Len(res.Rels(), 2)

		assert.Equal("abc", res.Get("str"))
		assert.Equal(123.456789, res.Get("float64"))
	})

	t.Run("partial resource arrays", func(t *testing.T) {
		assert := assert.New(t)

		payload := `{
			"id": "id1",
			"type": "mocktype4",
			"attributes": {
				"int8arr": [-32,-16, 0, 16, 32, 64, 127],
				"intarr": []
			}
		}`

		res, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.NoError(err)

		assert.Equal("id1", res.GetID())
		assert.Equal("mocktype4", res.GetType().Name)
		assert.Len(res.Attrs(), 2)
		assert.Len(res.Rels(), 0)

		assert.Equal([]int{}, res.Get("intarr"))
		assert.Equal([]int8{-32, -16, 0, 16, 32, 64, 127}, res.Get("int8arr"))
	})
}

func TestUnmarshalPartialResource_Invalid(t *testing.T) {
	// Setup
	typ, _ := BuildType(mocktype{})
	typ.NewFunc = func() Resource {
		return Wrap(&mocktype{})
	}
	typ4, err := BuildType(mockType4{})
	assert.NoError(t, err)

	schema := &Schema{Types: []Type{typ, typ4}}

	t.Run("invalid attribute", func(t *testing.T) {
		payload := `{
			"id": "abc123",
			"type": "mocktype",
			"attributes": {
				"int": "not an int"
			}
		}`

		_, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.EqualError(t, err, `jsonapi: invalid value "\"not an int\"" for field "int": `+
			`strconv.Atoi: parsing "\"not an int\"": invalid syntax`)

		var invalidFieldValueErr *InvalidFieldValueError
		assert.ErrorAs(t, err, &invalidFieldValueErr)
		assert.Equal(t, "mocktype", invalidFieldValueErr.Type)
		assert.Equal(t, "int", invalidFieldValueErr.Field)
		assert.Equal(t, "int", invalidFieldValueErr.FieldType)
		assert.Equal(t, `"not an int"`, invalidFieldValueErr.Value)

		var sourceErr srcError
		assert.ErrorAs(t, err, &sourceErr)

		src, isPtr := sourceErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/attributes/int", src)
	})

	t.Run("unknown attribute", func(t *testing.T) {
		payload := `{
			"id": "abc123",
			"type": "mocktype",
			"attributes": {
				"unknown": "abc"
			}
		}`

		_, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.EqualError(t, err, `jsonapi: field "unknown" does not exist in resource `+
			`type "mocktype"`)

		var unknownFieldErr *UnknownFieldError
		assert.ErrorAs(t, err, &unknownFieldErr)
		assert.Equal(t, "mocktype", unknownFieldErr.Type)
		assert.Equal(t, "unknown", unknownFieldErr.Field)
		assert.True(t, unknownFieldErr.IsUnknownAttr())
		assert.Equal(t, "", unknownFieldErr.RelPath())

		var sourceErr srcError
		assert.ErrorAs(t, err, &sourceErr)

		src, isPtr := sourceErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/attributes", src)
	})

	t.Run("invalid relationship", func(t *testing.T) {
		payload := `{
			"id": "abc123",
			"type": "mocktype",
			"relationships": {
				"to-1": {
					"data": 123
				}
			}
		}`

		_, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.EqualError(t, err, "json: cannot unmarshal number into Go value of type "+
			"jsonapi.Identifier")
		assert.ErrorIs(t, err, ErrInvalidPayload)

		var sourceErr srcError
		assert.ErrorAs(t, err, &sourceErr)

		src, isPtr := sourceErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/relationships/to-1", src)
	})

	t.Run("unknown relationship", func(t *testing.T) {
		payload := `{
			"id": "abc123",
			"type": "mocktype",
			"relationships": {
				"unknown": {
					"data": {
						"type": "mocktype",
						"data": "def"
					}
				}
			}
		}`

		_, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.EqualError(t, err, `jsonapi: field "unknown" does not exist in resource `+
			`type "mocktype"`)

		var unknownFieldErr *UnknownFieldError
		assert.ErrorAs(t, err, &unknownFieldErr)
		assert.Equal(t, "mocktype", unknownFieldErr.Type)
		assert.Equal(t, "unknown", unknownFieldErr.Field)
		assert.False(t, unknownFieldErr.IsUnknownAttr())
		assert.Equal(t, "", unknownFieldErr.RelPath())

		var sourceErr srcError
		assert.ErrorAs(t, err, &sourceErr)

		src, isPtr := sourceErr.Source()
		assert.True(t, isPtr)
		assert.Equal(t, "/relationships", src)
	})

	t.Run("invalid payload", func(t *testing.T) {
		payload := `{invalid}`

		_, err := UnmarshalPartialResource([]byte(payload), schema)
		assert.EqualError(t, err, "invalid character 'i' looking for beginning of object "+
			"key string")
	})
}

func TestEqual(t *testing.T) {
	assert := assert.New(t)

	now := time.Now()

	mt11 := Wrap(&mockType1{
		ID:             "mt1",
		Str:            "str",
		Int:            1,
		Int8:           2,
		Int16:          3,
		Int32:          4,
		Int64:          5,
		Uint:           6,
		Uint8:          7,
		Uint16:         8,
		Uint32:         9,
		Bool:           true,
		Time:           now,
		ToOne:          "a",
		ToOneFromOne:   "b",
		ToOneFromMany:  "c",
		ToMany:         []string{"a", "b", "c"},
		ToManyFromOne:  []string{"a", "b", "c"},
		ToManyFromMany: []string{"a", "b", "c"},
	})

	mt12 := Wrap(&mockType1{
		ID:             "mt2",
		Str:            "str",
		Int:            1,
		Int8:           2,
		Int16:          3,
		Int32:          4,
		Int64:          5,
		Uint:           6,
		Uint8:          7,
		Uint16:         8,
		Uint32:         9,
		Bool:           true,
		Time:           now,
		ToOne:          "a",
		ToOneFromOne:   "b",
		ToOneFromMany:  "c",
		ToMany:         []string{"a", "b", "c"},
		ToManyFromOne:  []string{"a", "b", "c"},
		ToManyFromMany: []string{"a", "b", "c"},
	})

	mt13 := Wrap(&mockType1{
		ID:             "mt3",
		Str:            "str",
		Int:            11,
		Int8:           12,
		Int16:          13,
		Int32:          14,
		Int64:          15,
		Uint:           16,
		Uint8:          17,
		Uint16:         18,
		Uint32:         19,
		Bool:           false,
		Time:           time.Now(),
		ToOne:          "d",
		ToOneFromOne:   "e",
		ToOneFromMany:  "f",
		ToMany:         []string{"d", "e", "f"},
		ToManyFromOne:  []string{"d", "e", "f"},
		ToManyFromMany: []string{"d", "e", "f"},
	})

	mt21 := Wrap(&mockType2{
		ID:             "mt1",
		StrPtr:         func() *string { v := "id"; return &v }(),
		IntPtr:         func() *int { v := int(1); return &v }(),
		Int8Ptr:        func() *int8 { v := int8(2); return &v }(),
		Int16Ptr:       func() *int16 { v := int16(3); return &v }(),
		Int32Ptr:       func() *int32 { v := int32(4); return &v }(),
		Int64Ptr:       func() *int64 { v := int64(5); return &v }(),
		UintPtr:        func() *uint { v := uint(6); return &v }(),
		Uint8Ptr:       func() *uint8 { v := uint8(7); return &v }(),
		Uint16Ptr:      func() *uint16 { v := uint16(8); return &v }(),
		Uint32Ptr:      func() *uint32 { v := uint32(9); return &v }(),
		BoolPtr:        func() *bool { v := true; return &v }(),
		TimePtr:        func() *time.Time { v := time.Now(); return &v }(),
		ToOneFromOne:   "a",
		ToOneFromMany:  "b",
		ToManyFromOne:  []string{"a", "b", "c"},
		ToManyFromMany: []string{"a", "b", "c"},
	})

	assert.True(Equal(mt11, mt11), "same instance")
	assert.True(Equal(mt11, mt12), "identical resources")
	assert.False(EqualStrict(mt11, mt12), "different IDs")
	assert.False(Equal(mt11, mt13), "different resources (same type)")
	assert.False(Equal(mt11, mt21), "different types")

	typ := mt11.GetType().Copy()
	sr1 := &SoftResource{Type: &typ}
	sr1.RemoveField("str")
	assert.False(Equal(mt11, sr1), "different number of attributes")

	typ = mt11.GetType().Copy()
	sr1 = &SoftResource{Type: &typ}

	for _, attr := range typ.Attrs {
		sr1.Set(attr.Name, mt11.Get(attr.Name))
	}

	for _, rel := range typ.Rels {
		if rel.ToOne {
			sr1.Set(rel.FromName, mt11.Get(rel.FromName).(string))
		} else {
			sr1.Set(rel.FromName, mt11.Get(rel.FromName).([]string))
		}
	}

	sr1.RemoveField("to-one")
	assert.False(Equal(mt11, sr1), "different number of relationships")

	sr1.AddRel(Rel{
		FromName: "to-one",
		ToOne:    false,
		ToType:   "mocktypes2",
	})
	assert.False(Equal(mt11, sr1), "different to-one property")

	sr1.RemoveField("to-one")
	sr1.AddRel(Rel{
		FromName: "to-one",
		ToOne:    true,
		ToType:   "mocktypes2",
	})
	sr1.Set("to-one", "b")
	assert.False(Equal(mt11, sr1), "different relationship value (to-one)")

	sr1.Set("to-one", "a")
	sr1.Set("to-many", []string{"d", "e", "f"})
	assert.False(Equal(mt11, sr1), "different relationship value (to-many)")

	// Comparing two nil values of different types
	sr3 := &SoftResource{}
	sr3.AddAttr(Attr{
		Name:     "nil",
		Type:     AttrTypeString,
		Nullable: true,
	})
	sr3.Set("nil", (*string)(nil))

	sr4 := &SoftResource{}
	sr4.AddAttr(Attr{
		Name:     "nil2",
		Type:     AttrTypeInt,
		Nullable: true,
	})
	sr3.Set("nil", (*int)(nil))

	assert.Equal(true, Equal(sr3, sr4))
}

func TestEqualStrict(t *testing.T) {
	assert := assert.New(t)

	sr1 := &SoftResource{}
	sr1.SetType(&Type{
		Name: "type",
	})

	sr2 := &SoftResource{}
	sr2.SetType(&Type{
		Name: "type",
	})

	// Same ID
	sr1.SetID("an-id")
	sr2.SetID("an-id")
	assert.True(Equal(sr1, sr2))
	assert.True(EqualStrict(sr1, sr2))

	// Different ID
	sr2.SetID("another-id")
	assert.True(Equal(sr1, sr2))
	assert.False(EqualStrict(sr1, sr2))
}
