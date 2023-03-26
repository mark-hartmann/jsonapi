package jsonapi_test

import (
	"testing"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	assert := assert.New(t)

	err := Check("not a struct")
	assert.EqualError(err, "jsonapi: not a struct")

	err = Check(emptyIDAPItag{})
	assert.EqualError(err, "jsonapi: ID field's api tag is empty")

	err = Check(invalidAttributeType{})
	assert.EqualError(
		err,
		"jsonapi: attribute \"Attr\" of type \"typename\" is of unsupported type",
	)

	err = Check(invalidRelAPITag{})
	assert.EqualError(
		err,
		"jsonapi: api tag of relationship \"Rel\" of struct \"invalidRelAPITag\" is invalid",
	)

	err = Check(invalidReType{})
	assert.EqualError(
		err,
		"jsonapi: relationship \"Rel\" of type \"typename\" is not string or []string",
	)

	err = Check(mockType4{})
	assert.NoError(err)

	err = Check(mockType5{})
	assert.NoError(err)
}

func TestBuildType(t *testing.T) {
	assert := assert.New(t)

	assert.Panics(func() {
		MustBuildType("invalid")
	})

	mock := mockType1{
		ID:    "abc13",
		Str:   "string",
		Int:   -42,
		Uint8: 12,
	}
	typ, err := BuildType(mock)
	assert.NoError(err)
	assert.Equal(true, Equal(Wrap(&mockType1{}), typ.New()))

	// Build type from pointer to struct
	typ, err = BuildType(&mock)
	assert.NoError(err)
	assert.Equal(true, Equal(Wrap(&mockType1{}), typ.New()))

	typ, err = BuildType(&mockType4{})
	assert.NoError(err)
	assert.True(Equal(Wrap(&mockType4{}), typ.New()))

	typ, err = BuildType(&mockType6{})
	assert.NoError(err)
	assert.True(Equal(Wrap(&mockType6{}), typ.New()))

	// Build from invalid struct
	_, err = BuildType(invalidRelAPITag{})
	assert.Error(err)
}

type emptyIDAPItag struct {
	ID string `json:"id"`
}

type invalidAttributeType struct {
	ID   string `json:"id" api:"typename"`
	Attr error  `json:"attr" api:"attr"`
}

type invalidRelAPITag struct {
	ID  string `json:"id" api:"typename"`
	Rel string `json:"rel" api:"rel,but,it,is,invalid"`
}

type invalidReType struct {
	ID  string `json:"id" api:"typename"`
	Rel int    `json:"rel" api:"rel,target,reverse"`
}

func TestReduceRels(t *testing.T) {
	tests := map[string]struct {
		in     []Rel
		result []Rel
	}{
		"no change": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
				{FromName: "c->d", FromType: "c", ToType: "d"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
				{FromName: "c->d", FromType: "c", ToType: "d"},
			},
		},
		"very simple cycle": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->a", FromType: "b", ToType: "a"},
			},
			result: []Rel{},
		},
		"simple cycle": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
				{FromName: "c->f", FromType: "c", ToType: "f"},
				{FromName: "f->c", FromType: "f", ToType: "c"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
			},
		},
		"full blown loop": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"}, // 0
				{FromName: "b->a", FromType: "b", ToType: "a"}, // 1
				{FromName: "a->b", FromType: "a", ToType: "b"}, // 2
				{FromName: "b->a", FromType: "b", ToType: "a"}, // 3
				{FromName: "a->b", FromType: "a", ToType: "b"}, // 4
				{FromName: "b->a", FromType: "b", ToType: "a"}, // 5
			},
			result: []Rel{},
		},
		"full blown loop 2": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->a", FromType: "b", ToType: "a"},
				{FromName: "a->b", FromType: "a", ToType: "b"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
			},
		},
		"multiple redundant connections": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
				{FromName: "c->f", FromType: "c", ToType: "f"},
				{FromName: "f->b", FromType: "f", ToType: "b"},
				{FromName: "b->x", FromType: "b", ToType: "x"},
				{FromName: "x->y", FromType: "x", ToType: "y"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->x", FromType: "b", ToType: "x"},
				{FromName: "x->y", FromType: "x", ToType: "y"},
			},
		},
		"self-rel only": {
			in: []Rel{
				{FromName: "a->a", FromType: "a", ToType: "a"},
			},
			result: []Rel{},
		},
		"self-rel first": {
			in: []Rel{
				{FromName: "a->a", FromType: "a", ToType: "a"},
				{FromName: "a->b", FromType: "a", ToType: "b"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
			},
		},
		"self-rel between": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->b", FromType: "b", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
			},
		},
		"self-rel last": {
			in: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
				{FromName: "c->c", FromType: "c", ToType: "c"},
			},
			result: []Rel{
				{FromName: "a->b", FromType: "a", ToType: "b"},
				{FromName: "b->c", FromType: "b", ToType: "c"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.result, ReduceRels(test.in))
		})
	}
}
