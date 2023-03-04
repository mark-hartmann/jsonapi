package jsonapi_test

import (
	"net/url"
	"testing"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestSimpleURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedURL   SimpleURL
		expectedError error
	}{
		{
			name: "empty url",
		}, {
			name: "empty path",
			url:  `https://api.example.com`,
		}, {
			name: "collection",
			url:  `https://api.example.com/type`,
			expectedURL: SimpleURL{
				Fragments: []string{"type"},
				Route:     "/type",
			},
		}, {
			name: "resource",
			url:  `https://api.example.com/type/id`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id"},
				Route:     "/type/:id",
			},
		}, {
			name: "relationship",
			url:  `https://api.example.com/type/id/rel`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",
			},
		}, {
			name: "self relationship",
			url:  `https://api.example.com/type/id/relationships/rel`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "relationships", "rel"},
				Route:     "/type/:id/relationships/rel",
			},
		}, {
			name: "fields, sort, pagination, include",
			url: `https://api.example.com/type
				?fields[type]=attr1,attr2,rel1
				&fields[type2]=attr3,attr4,rel2,rel3
				&fields[type3]=attr5,attr6,rel4
				&fields[type4]=attr7,rel5,rel6
				&sort=attr2,-attr1
				&page[number]=1
				&page[size]=20
				&include=type2.type3,type4
			`,
			expectedURL: SimpleURL{
				Fragments: []string{"type"},
				Route:     "/type",

				Fields: map[string][]string{
					"type":  {"attr1", "attr2", "rel1"},
					"type2": {"attr3", "attr4", "rel2", "rel3"},
					"type3": {"attr5", "attr6", "rel4"},
					"type4": {"attr7", "rel5", "rel6"},
				},
				SortingRules: []string{
					"attr2",
					"-attr1",
				},
				Page: map[string]string{
					"size":   "20",
					"number": "1",
				},
				Include: []string{
					"type2.type3",
					"type4",
				},
			},
		}, {
			name: "duplicated and overlapping includes",
			url:  `https://api.example.com/type?include=type2.type3,type4,type2,type3`,
			expectedURL: SimpleURL{
				Fragments: []string{"type"},
				Route:     "/type",

				Include: []string{
					"type2.type3",
					"type4",
					"type2",
					"type3",
				},
			},
		}, {
			name: "fields in separate definitions",
			url: `https://api.example.com/type
				?fields[type]=attr1,attr2,rel1
				&fields[type2]=attr3,attr4,rel2,rel3
				&fields[type]=attr3,attr4,rel2
			`,
			expectedURL: SimpleURL{
				Fragments: []string{"type"},
				Route:     "/type",

				Fields: map[string][]string{
					"type":  {"attr1", "attr2", "rel1", "attr3", "attr4", "rel2"},
					"type2": {"attr3", "attr4", "rel2", "rel3"},
				},
			},
		}, {
			name: "duplicate fields",
			url: `https://api.example.com/type/id/rel
				?fields[type]=attr1,attr2,attr1
				&fields[type2]=rel3,attr4`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Fields: map[string][]string{
					"type":  {"attr1", "attr2", "attr1"},
					"type2": {"rel3", "attr4"},
				},
			},
		}, {
			name: "filter label",
			url:  `https://api.example.com/type/id/rel?filter=label`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Filter: map[string][]string{
					"filter": {"label"},
				},
			},
		}, {
			name: "negative page size",
			url:  `https://api.example.com/type/id/rel?page[size]=-1`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Page: map[string]string{
					"size": "-1",
				},
			},
		}, {
			name: "empty page query param",
			url:  `https://api.example.com/type/id/rel?page[test]&page[size]=40`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Page: map[string]string{
					"test": "",
					"size": "40",
				},
			},
		}, {
			name: "unknown parameter",
			url:  `https://api.example.com/type/id/rel?unknownparam=somevalue`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Params: map[string][]string{
					"unknownparam": {"somevalue"},
				},
			},
		}, {
			name: "filter query",
			url:  `https://api.example.com/type/id/rel?filter={"f": "field","o": "=","v": "abc"}`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Filter: map[string][]string{
					"filter": {
						"{\"f\":\"field\",\"o\":\"=\",\"v\":\"abc\"}",
					},
				},
			},
		}, {
			name: "filter query",
			url:  `https://api.example.com/type/id/rel?filter={"thisis:afilter"}`,
			expectedURL: SimpleURL{
				Fragments: []string{"type", "id", "rel"},
				Route:     "/type/:id/rel",

				Filter: map[string][]string{
					"filter": {"{\"thisis:afilter\"}"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u, err := url.Parse(makeOneLineNoSpaces(test.url))
			assert.NoError(t, err)

			su, err := NewSimpleURL(u)

			if test.expectedError != nil {
				jaErr := test.expectedError.(Error)
				jaErr.ID = ""
				test.expectedError = jaErr
			}

			if err != nil {
				jaErr := err.(Error)
				jaErr.ID = ""
				err = jaErr
			}

			assert.Equal(t, test.expectedURL, su)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestSimpleURLPath(t *testing.T) {
	su := &SimpleURL{Fragments: []string{}}
	assert.Equal(t, "", su.Path())

	su = &SimpleURL{Fragments: []string{"a", "b", "c"}}
	assert.Equal(t, "a/b/c", su.Path())
}
