package jsonapi_test

import (
	"net/url"
	"testing"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	// Schema
	schema := newMockSchema()

	tests := map[string]struct {
		url           string
		expectedURL   URL
		expectedError bool
	}{
		"empty": {
			url:           ``,
			expectedError: true,
		},
		"empty path": {
			url:           `https://example.com`,
			expectedError: true,
		},
		"type not found": {
			url:           "/mocktypes99",
			expectedError: true,
		},
		"relationship not found": {
			url:           "/mocktypes1/abc/relnotfound",
			expectedError: true,
		},
		"bad params": {
			url:           `/mocktypes1?fields[invalid]=attr1,attr2`,
			expectedError: true,
		},
		"invalid raw url": {
			url:           "%z",
			expectedError: true,
		},
		"collection name only": {
			url: `mocktypes1`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1",
				},
				Route:   "/mocktypes1",
				ResType: "mocktypes1",
				IsCol:   true,
			},
		},
		"string value page param": {
			url: `/mocktypes1/abc123?page[size]=valid`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "abc123",
				},
				Route:   "/mocktypes1/:id",
				ResType: "mocktypes1",
				ResID:   "abc123",
			},
		},
		"invalid simple url": {
			url: `/mocktypes1/abc123?page=no-page-param`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "abc123",
				},
				Route:   "/mocktypes1/:id",
				ResType: "mocktypes1",
				ResID:   "abc123",
			},
		},
		"full url for collection": {
			url: `https://api.example.com/mocktypes1`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1",
				},
				Route:   "/mocktypes1",
				ResType: "mocktypes1",
				IsCol:   true,
			},
		},
		"full url for resource": {
			url: `https://example.com/mocktypes1/mc1-1`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "mc1-1",
				},
				Route:   "/mocktypes1/:id",
				ResType: "mocktypes1",
				ResID:   "mc1-1",
			},
		},
		"full url for related relationship": {
			url: `https://example.com/mocktypes1/mc1-1/to-one`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "mc1-1", "to-one",
				},
				Route: "/mocktypes1/:id/to-one",
				BelongsToFilter: BelongsToFilter{
					Type: "mocktypes1",
					ID:   "mc1-1",
					Name: "to-one",
				},
				ResType: "mocktypes2",
				RelKind: "related",
				Rel: Rel{
					FromName: "to-one",
					FromType: "mocktypes1",
					ToOne:    true,
					ToName:   "",
					ToType:   "mocktypes2",
					FromOne:  false,
				},
			},
		},
		"full url for self relationships": {
			url: `https://example.com/mocktypes1/mc1-1/relationships/to-many-from-one`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "mc1-1", "relationships", "to-many-from-one",
				},
				Route: "/mocktypes1/:id/relationships/to-many-from-one",
				BelongsToFilter: BelongsToFilter{
					Type:   "mocktypes1",
					ID:     "mc1-1",
					Name:   "to-many-from-one",
					ToName: "to-one-from-many",
				},
				ResType: "mocktypes2",
				RelKind: "self",
				Rel: Rel{
					FromName: "to-many-from-one",
					FromType: "mocktypes1",
					ToOne:    false,
					ToName:   "to-one-from-many",
					ToType:   "mocktypes2",
					FromOne:  true,
				},
				IsCol: true,
			},
		},
		"path for self relationship": {
			url: `/mocktypes1/mc1-1/relationships/to-many-from-one`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "mc1-1", "relationships", "to-many-from-one",
				},
				Route: "/mocktypes1/:id/relationships/to-many-from-one",
				BelongsToFilter: BelongsToFilter{
					Type:   "mocktypes1",
					ID:     "mc1-1",
					Name:   "to-many-from-one",
					ToName: "to-one-from-many",
				},
				ResType: "mocktypes2",
				RelKind: "self",
				Rel: Rel{
					FromName: "to-many-from-one",
					FromType: "mocktypes1",
					ToOne:    false,
					ToName:   "to-one-from-many",
					ToType:   "mocktypes2",
					FromOne:  true,
				},
				IsCol: true,
			},
		},
		"path for self relationship with params": {
			url: `/mocktypes1/mc1-1/relationships/to-many-from-one
?fields[mocktypes2]=boolptr%2Cint8ptr`,
			expectedURL: URL{
				Fragments: []string{
					"mocktypes1", "mc1-1", "relationships", "to-many-from-one",
				},
				Route: "/mocktypes1/:id/relationships/to-many-from-one",
				BelongsToFilter: BelongsToFilter{
					Type:   "mocktypes1",
					ID:     "mc1-1",
					Name:   "to-many-from-one",
					ToName: "to-one-from-many",
				},
				ResType: "mocktypes2",
				RelKind: "self",
				Rel: Rel{
					FromName: "to-many-from-one",
					FromType: "mocktypes1",
					ToOne:    false,
					ToName:   "to-one-from-many",
					ToType:   "mocktypes2",
					FromOne:  true,
				},
				IsCol: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u, err := NewURLFromRaw(schema, makeOneLineNoSpaces(test.url))

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				u.Params = nil
				assert.Equal(t, test.expectedURL, *u)
			}
		})
	}
}

func TestParseParams(t *testing.T) {
	// Schema
	schema := newMockSchema()
	mockTypes1 := schema.GetType("mocktypes1")
	mockTypes2 := schema.GetType("mocktypes2")

	tests := []struct {
		name           string
		url            string
		colType        string
		expectedParams Params
		expectedError  bool
	}{
		{
			name: "slash only",
			url:  `/`,
		}, {
			name: "question mark",
			url:  `?`,
		}, {
			name: "sort, pagination and offspec query params",
			url: `/mocktypes1?fields[mocktypes1]=bool,str,uint8&sort=str,-bool
&fields[mocktypes2]=intptr,strptr&page[number]=20&page[size]=50&include=to-many-from-one
&foo=foo&foo=bar&foo=baz&test[a]=b&test[b]=c&test[b]=d`,
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
		}, {
			name: "include, sort, pagination in multiple parts",
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
		}, {
			name: "sort param with many escaped commas",
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
		}, {
			name: "sort param with many unescaped commas",
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
		}, {
			name:    "filter label",
			url:     `?filter=label`,
			colType: "mocktypes1",
			expectedParams: Params{
				Filter: map[string][]string{
					"filter": {"label"},
				},
			},
		}, {
			name:    "multiple filter labels",
			url:     `?filter=label&filter=label2&filter[foo]=bar&filter[10%257]=3`,
			colType: "mocktypes1",
			expectedParams: Params{
				Filter: map[string][]string{
					"filter":       {"label", "label2"},
					"filter[foo]":  {"bar"},
					"filter[10%7]": {"3"},
				},
			},
		}, {
			name:    "sorting rules without id",
			url:     `/mocktypes1?sort=str,-int`,
			colType: "mocktypes1",
			expectedParams: Params{
				SortRules: []SortRule{
					{Name: "str"},
					{Name: "int", Desc: true},
				},
			},
		}, {
			name: "sorting rules with id",
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,int,int16,int32,int64,int8,str,time,to-many,
to-many-from-many,to-many-from-one,to-one,to-one-from-many,to-one-from-one,uint,uint16,uint32,
uint64,uint8
				&sort=str,-int,id,-to-one-from-one.int16ptr,to-one-from-one.to-one-from-many.str
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
		}, {
			name: "duplicate fields in sparse fieldset",
			url: `/mocktypes1?fields[mocktypes1]=bool,str,uint8&foo=bar
&fields[mocktypes1]=some-unknown-field`,
			expectedError: true,
		}, {
			name: "inclusion of unknown relationship",
			url: `
				?include=some-unknown-relationship
			`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name: "invalid sort path (to-many)",
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,int,int16&sort=str,-int,id,-to-many-from-one.int16ptr
			`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name: "invalid sort path (unknown relationship)",
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool,int,int16&sort=str,-int,id,-doesnotexist.int16ptr
			`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name:          "unknown sort field",
			url:           `/mocktypes1?sort=doesnotexist`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name:          "unknown sort field (deep)",
			url:           `/mocktypes1?sort=to-one-from-one.doesnotexist`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name: "fields with duplicates",
			url: `
				/mocktypes1
				?fields[mocktypes1]=str,int16,bool,str
			`,
			colType: "mocktypes1",
			expectedParams: Params{
				Fields: map[string][]string{
					"mocktypes1": {"bool", "int16", "str"},
				},
			},
		}, {
			name: "fields with id",
			url: `
				/mocktypes1
				?fields[mocktypes1]=str,id
			`,
			colType: "mocktypes1",
			expectedParams: Params{
				Fields: map[string][]string{
					"mocktypes1": {"id", "str"},
				},
			},
		}, {
			name: "conflicting sort rules",
			url: `
				/mocktypes1?sort=str,int,-str,int8
			`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name: "invalid sort rule",
			url: `
				/mocktypes1?sort=-
			`,
			colType:       "mocktypes1",
			expectedError: true,
		}, {
			name: "conflicting sort rules relationship path",
			url: `
				/mocktypes1?sort=to-one-from-one.strptr,-to-one-from-one.strptr
			`,
			colType:       "mocktypes1",
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u, err := url.Parse(makeOneLineNoSpaces(test.url))
			assert.NoError(t, err, test.name)

			su, err := NewSimpleURL(u)
			assert.NoError(t, err, test.name)

			params, err := NewParams(schema, su, test.colType)

			// Set Attrs and Rels if mentioned in Fields
			for colType, fields := range test.expectedParams.Fields {
				for _, field := range fields {
					if typ := schema.GetType(colType); typ.Name != "" {
						if _, ok := typ.Attrs[field]; ok {
							if len(test.expectedParams.Attrs) == 0 {
								test.expectedParams.Attrs = map[string][]Attr{}
							}

							test.expectedParams.Attrs[colType] = append(
								test.expectedParams.Attrs[colType],
								typ.Attrs[field],
							)
						} else if typ := schema.GetType(colType); typ.Name != "" {
							if _, ok := typ.Rels[field]; ok {
								if len(test.expectedParams.Rels) == 0 {
									test.expectedParams.Rels = map[string][]Rel{}
								}

								test.expectedParams.Rels[colType] = append(
									test.expectedParams.Rels[colType],
									typ.Rels[field],
								)
							}
						}
					}
				}
			}

			if test.expectedError {
				assert.Error(t, err, test.name)
			} else {
				assert.NoError(t, err, test.name)
				assert.Equal(t, test.expectedParams, *params, test.name)
			}
		})
	}
}

func TestURLEscaping(t *testing.T) {
	assert := assert.New(t)

	schema := newMockSchema()

	tests := []struct {
		url       string
		escaped   string
		unescaped string
	}{
		{
			url: `
				/mocktypes1
				?fields[mocktypes1]=bool%2Cint8
				&page[number]=2
				&page[size]=10
				&page[abc]
				&filter=a_label
				&sort=bool,int,int16,uint8,id,-to-one.boolptr
			`,
			escaped: `
				/mocktypes1
				?fields%5Bmocktypes1%5D=bool%2Cint8
				&filter=a_label
				&page%5Babc%5D=
				&page%5Bnumber%5D=2
				&page%5Bsize%5D=10
				&sort=bool%2Cint%2Cint16%2Cuint8%2Cid%2C-to-one.boolptr
				`,
			unescaped: `
				/mocktypes1
				?fields[mocktypes1]=bool,int8
				&filter=a_label
				&page[abc]=
				&page[number]=2
				&page[size]=10
				&sort=bool,int,int16,uint8,id,-to-one.boolptr
			`,
		},
	}

	for _, test := range tests {
		url, err := NewURLFromRaw(schema, makeOneLineNoSpaces(test.url))
		assert.NoError(err)
		assert.Equal(
			makeOneLineNoSpaces(test.escaped),
			url.String(),
		)
		assert.Equal(
			makeOneLineNoSpaces(test.unescaped),
			url.UnescapedString(),
		)
	}
}

func TestURLString(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		raw, expected string
	}{
		"simple resource collection": {
			raw:      "/mocktypes1",
			expected: "/mocktypes1",
		},
		"overlapping inclusions and sorting rules": {
			raw: `/mocktypes1?include=to-many-from-one.to-one-from-many&sort=uint8&include=
		to-many-from-one&sort=-str`,
			expected: "/mocktypes1?include=to-many-from-one.to-one-from-many&sort=uint8,-str",
		},
		"complete example": {
			raw: `
		/mocktypes1
		?fields[mocktypes1]=bool,int,int16,int32,int64,int8,str,time,to-many,to-many-from-many
		&include=
			to-many-from-one.to-one-from-many.to-one.to-many-from-many%2C
			to-one-from-one.to-many-from-many
		&sort=str,%2C%2C-bool
		&fields[mocktypes2]=boolptr,int16ptr,int32ptr
		&page[number]=3
		&sort=uint8,to-one-from-one.to-one-from-one.str,-to-one-from-one.boolptr
		&include=
			to-many-from-one,
			to-many-from-many
		&page[size]=50
		&filter={"f":"str","o":"=","v":"abc"}
	`,
			expected: `
		/mocktypes1
		?fields[mocktypes1]=bool,int,int16,int32,int64,int8,str,time,to-many,to-many-from-many
		&fields[mocktypes2]=boolptr,int16ptr,int32ptr
		&include=to-many-from-many,to-many-from-one.to-one-from-many.to-one.to-many-from-many,
		to-one-from-one.to-many-from-many
		&filter={"f":"str","o":"=","v":"abc"}
		&page[number]=3
		&page[size]=50
		&sort=str,-bool,uint8,str,-to-one-from-one.boolptr
		`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u, err := NewURLFromRaw(newMockSchema(), makeOneLineNoSpaces(test.raw))

			assert.NoError(err)
			assert.Equal(makeOneLineNoSpaces(test.expected), u.UnescapedString())
		})
	}
}
