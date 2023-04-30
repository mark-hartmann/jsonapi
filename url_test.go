package jsonapi_test

import (
	"strings"
	"testing"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestNewURL(t *testing.T) {
	// Schema
	schema := newMockSchema()

	t.Run("errors", func(t *testing.T) {
		invalidTests := map[string]struct {
			url string
			err string
		}{
			"empty": {
				url: ``,
				err: `jsonapi: empty path`,
			},
			"empty path": {
				url: `https://example.com`,
				err: `jsonapi: empty path`,
			},
			"type not found": {
				url: "/mocktypes99",
				err: `jsonapi: resource type "mocktypes99" does not exist`,
			},
			"relationship not found": {
				url: "/mocktypes1/abc/relnotfound",
				err: `jsonapi: field "relnotfound" does not exist in resource type "mocktypes1"`,
			},
			"bad params": {
				url: `/mocktypes1?fields[invalid]=attr1,attr2`,
				err: "" +
					"jsonapi: failed to create jsonapi.Params: " +
					`jsonapi: resource type "invalid" does not exist`,
			},
			"invalid raw url": {
				url: "%z",
				err: `jsonapi: failed to parse url.URL: parse "%z": invalid URL escape "%z"`,
			},
			"page param (no collection)": {
				url: `/mocktypes1/abc123?page[size]=valid`,
				err: `jsonapi: illegal query parameter "page"`,
			},
			"filter param (no collection)": {
				url: `/mocktypes1/abc123?filter=is-foo`,
				err: `jsonapi: illegal query parameter "filter"`,
			},
			"sort param (no collection)": {
				url: `/mocktypes1/abc123?sort=-str`,
				err: `jsonapi: illegal query parameter "sort"`,
			},
		}

		for name, test := range invalidTests {
			t.Run(name, func(t *testing.T) {
				_, err := NewURLFromRaw(schema, makeOneLineNoSpaces(test.url))
				assert.Error(t, err)
				assert.EqualError(t, err, test.err)
			})
		}

		t.Run("invalid url string", func(t *testing.T) {
			_, err := NewURLFromRaw(schema, "1http://foo.com")
			assert.Error(t, err)

			// Since go 1.14, the Error function uses the fmt package instead of basic
			// string concatenations, quoting the url we tried to parse. Since this
			// library support go 1.13, we, for now, simply check the prefix.
			//
			// todo: Replace with appropriate assertions after the supported go version
			//       is updated (#16)
			assert.True(t, strings.HasPrefix(err.Error(), "jsonapi: failed to parse url.URL"))
		})

		t.Run("empty url", func(t *testing.T) {
			_, err := NewURLFromRaw(schema, "")
			assert.Error(t, err)
			assert.EqualError(t, err, "jsonapi: empty path")

			var pErr interface{ InPath() bool }
			assert.ErrorAs(t, err, &pErr)
			assert.True(t, pErr.InPath())
		})

		t.Run("unknown type in url", func(t *testing.T) {
			_, err := NewURLFromRaw(schema, makeOneLineNoSpaces("/mocktypes99"))
			assert.Error(t, err)

			var unknownTypeErr *UnknownTypeError
			assert.ErrorAs(t, err, &unknownTypeErr)
			assert.Equal(t, "mocktypes99", unknownTypeErr.Type)
			assert.True(t, unknownTypeErr.InPath())
		})

		t.Run("unknown rel in url", func(t *testing.T) {
			_, err := NewURLFromRaw(schema, makeOneLineNoSpaces("/mocktypes1/abc/relnotfound"))
			assert.Error(t, err)

			var unknownFieldErr *UnknownFieldError
			assert.ErrorAs(t, err, &unknownFieldErr)
			assert.Equal(t, "mocktypes1", unknownFieldErr.Type)
			assert.Equal(t, "relnotfound", unknownFieldErr.Field)
			assert.False(t, unknownFieldErr.IsAttr())
			assert.True(t, unknownFieldErr.InPath())
			assert.Equal(t, "", unknownFieldErr.RelPath())
		})

		t.Run("collection param on single resource", func(t *testing.T) {
			_, err := NewURLFromRaw(schema, makeOneLineNoSpaces("/mocktypes1/abc?sort=int8"))
			assert.Error(t, err)

			var illegalParameterErr *IllegalParameterError
			assert.ErrorAs(t, err, &illegalParameterErr)
			assert.True(t, illegalParameterErr.IsResource())

			src, isPtr := illegalParameterErr.Source()
			assert.Equal(t, "sort", src)
			assert.False(t, isPtr)
		})
	})

	tests := map[string]struct {
		url         string
		expectedURL URL
	}{
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
			assert.NoError(t, err)

			// Set u.Params nil because we only validate the jsonapi.URL itself.
			u.Params = nil
			assert.Equal(t, test.expectedURL, *u)
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
