package jsonapi_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/mark-hartmann/jsonapi"
)

func TestNewParams(t *testing.T) {
	// Schema
	schema := newMockSchema()
	mockTypes1 := schema.GetType("mocktypes1")
	mockTypes2 := schema.GetType("mocktypes2")

	tests := map[string]struct {
		url            string
		colType        string
		expectedParams Params
		expectedError  bool
	}{
		"slash only": {
			url: `/`,
		},
		"question mark": {
			url: `?`,
		},
		"sort, pagination and off-spec query params": {
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,str,uint8
				&sort=str,-bool
				&fields[mocktypes2]=intptr,strptr
				&page[number]=20
				&page[size]=50
				&include=to-many-from-one
				&foo=foo
				&foo=bar
				&foo=baz
				&test[a]=b
				&test[b]=c
				&test[b]=d`,
			colType: "mocktypes1",
			expectedParams: Params{
				// mocktypes1 was requested, but without sparse fieldset. Since no relationship
				// was requested to be included, mocktypes2 does not appear in the maps.
				Fields: map[string][]string{
					"mocktypes1": {"bool", "str", "uint8"},
					"mocktypes2": {"intptr", "strptr"},
				},
				// SportingRules does not contain to-many-from-one because it's a relationship and
				// not an attribute.
				SortRules: []SortRule{
					{Name: "str"},
					{Name: "bool", Desc: true},
				},
				Page: map[string]string{
					"number": "20",
					"size":   "50",
				},
				Include: [][]Rel{
					{
						mockTypes1.Rels["to-many-from-one"],
					},
				},
				Params: map[string][]string{
					"foo":     {"foo", "bar", "baz"},
					"test[a]": {"b"},
					"test[b]": {"c", "d"},
				},
			},
		},
		"include, sort, pagination in multiple parts": {
			url: `
				?include=
					to-many-from-one.to-one-from-many.to-one.to-many-from-many,
					to-one-from-one.to-many-from-many
				&sort=to-many,str,,-bool
				&page[number]=3
				&sort=uint8
				&include=
					to-many-from-one,
					to-many-from-many
				&page[size]=50
			`,
			colType: "mocktypes1",
			expectedParams: Params{
				Page: map[string]string{"size": "50", "number": "3"},
				Include: [][]Rel{
					{
						mockTypes1.Rels["to-many-from-many"],
					},
					{
						mockTypes1.Rels["to-many-from-one"],
						mockTypes2.Rels["to-one-from-many"],
						mockTypes1.Rels["to-one"],
						mockTypes2.Rels["to-many-from-many"],
					},
					{
						mockTypes1.Rels["to-one-from-one"],
						mockTypes2.Rels["to-many-from-many"],
					},
				},
			},
		},
		"sort param with many escaped commas": {
			url: `
				?include=
					to-many-from-one.to-one-from-many.to-one.to-many-from-many%2C
					to-one-from-one.to-many-from-many
				&sort=to-many%2Cstr,%2C%2C-bool
				&page[number]=3
				&sort=uint8,-to-one-from-one.int16Ptr
				&include=
					to-many-from-one,
					to-many-from-many
				&page[size]=50
			`,
			colType: "mocktypes1",
			expectedParams: Params{
				Page: map[string]string{"size": "50", "number": "3"},
				Include: [][]Rel{
					{
						mockTypes1.Rels["to-many-from-many"],
					},
					{
						mockTypes1.Rels["to-many-from-one"],
						mockTypes2.Rels["to-one-from-many"],
						mockTypes1.Rels["to-one"],
						mockTypes2.Rels["to-many-from-many"],
					},
					{
						mockTypes1.Rels["to-one-from-one"],
						mockTypes2.Rels["to-many-from-many"],
					},
				},
			},
		},
		"sort param with many unescaped commas": {
			url: `
				?include=
					to-many-from-one.to-one-from-many
				&sort=to-many,str,,,-bool
				&sort=uint8
				&include=
					to-many-from-many.
					to-many-from-one,
				&page[number]=110
				&page[size]=90
			`,
			colType: "mocktypes1",
			expectedParams: Params{
				Page: map[string]string{"size": "90", "number": "110"},
				Include: [][]Rel{
					{
						mockTypes1.Rels["to-many-from-many"],
						mockTypes2.Rels["to-many-from-one"],
					},
					{
						mockTypes1.Rels["to-many-from-one"],
						mockTypes2.Rels["to-one-from-many"],
					},
				},
			},
		},
		"filter label": {
			url:     `?filter=label`,
			colType: "mocktypes1",
			expectedParams: Params{
				Filter: map[string][]string{
					"filter": {"label"},
				},
			},
		},
		"multiple filter labels": {
			url:     `?filter=label&filter=label2&filter[foo]=bar&filter[10%257]=3`,
			colType: "mocktypes1",
			expectedParams: Params{
				Filter: map[string][]string{
					"filter":       {"label", "label2"},
					"filter[foo]":  {"bar"},
					"filter[10%7]": {"3"},
				},
			},
		},
		"sorting rules without id": {
			url:     `/mocktypes1?sort=str,-int`,
			colType: "mocktypes1",
			expectedParams: Params{
				SortRules: []SortRule{
					{Name: "str"},
					{Name: "int", Desc: true},
				},
			},
		},
		"sorting rules with id": {
			url: `
				/mocktypes1
				?fields[mocktypes1]=
					bool,int,int16,int32,int64,int8,str,time,
					to-many,to-many-from-many,to-many-from-one,to-one,
					to-one-from-many,to-one-from-one,uint,uint16,uint32,
					uint64,uint8
				&sort=
					str,-int,id,
					-to-one-from-one.int16ptr,
					to-one-from-one.to-one-from-many.str
			`,
			colType: "mocktypes1",
			expectedParams: Params{
				Fields: map[string][]string{
					"mocktypes1": mockTypes1.Fields(),
				},
				SortRules: []SortRule{
					{Name: "str"},
					{Name: "int", Desc: true},
					{Name: "id"},
					{
						Path: []Rel{
							mockTypes1.Rels["to-one-from-one"],
						},
						Name: "int16ptr",
						Desc: true,
					},
					{Name: "str"},
				},
			},
		},
		"duplicate fields in sparse fieldset": {
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,str,uint8
				&foo=bar
				&fields[mocktypes1]=some-unknown-field`,
			expectedError: true,
		},
		"inclusion of unknown relationship": {
			url:           `?include=some-unknown-relationship`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"invalid sort path (to-many)": {
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,int,int16
				&sort=str,-int,id,-to-many-from-one.int16ptr`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"invalid sort path (unknown relationship)": {
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,int,int16
				&sort=str,-int,id,-doesnotexist.int16ptr
			`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"unknown sort field": {
			url:           `/mocktypes1?sort=doesnotexist`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"unknown sort field (deep)": {
			url:           `/mocktypes1?sort=to-one-from-one.doesnotexist`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"fields with duplicates": {
			url:     `/mocktypes1?fields[mocktypes1]=str,int16,bool,str`,
			colType: "mocktypes1",
			expectedParams: Params{
				Fields: map[string][]string{
					"mocktypes1": {"bool", "int16", "str"},
				},
			},
		},
		"fields with id": {
			url:     `/mocktypes1?fields[mocktypes1]=str,id`,
			colType: "mocktypes1",
			expectedParams: Params{
				Fields: map[string][]string{
					"mocktypes1": {"id", "str"},
				},
			},
		},
		"conflicting sort rules": {
			url:           `/mocktypes1?sort=str,int,-str,int8`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"invalid sort rule": {
			url:           `/mocktypes1?sort=-`,
			colType:       "mocktypes1",
			expectedError: true,
		},
		"conflicting sort rules relationship path": {
			url:           `/mocktypes1?sort=to-one-from-one.strptr,-to-one-from-one.strptr`,
			colType:       "mocktypes1",
			expectedError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u, err := url.Parse(makeOneLineNoSpaces(test.url))
			assert.NoError(t, err)

			su, err := NewSimpleURL(u)
			assert.NoError(t, err)

			params, err := NewParams(schema, su, test.colType)

			// So that one does not have to set the Attr and Rel maps manually, we set it up
			// like this.
			test.expectedParams.Attrs, test.expectedParams.Rels = getExpectedAttrsAndRels(
				schema, test.expectedParams.Fields)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedParams, *params)
			}
		})
	}
}

func getExpectedAttrsAndRels(schema *Schema, fieldMap map[string][]string) (
	attrs map[string][]Attr, rels map[string][]Rel) {
	for resType, fields := range fieldMap {
		typ := schema.GetType(resType)
		if typ.Name == "" {
			continue
		}

		for _, field := range fields {
			if attr, ok := typ.Attrs[field]; ok {
				if len(attrs) == 0 {
					attrs = map[string][]Attr{}
				}

				attrs[typ.Name] = append(attrs[typ.Name], attr)
			} else if rel, ok := typ.Rels[field]; ok {
				if len(rels) == 0 {
					rels = map[string][]Rel{}
				}

				rels[typ.Name] = append(rels[typ.Name], rel)
			}
		}
	}

	return
}
