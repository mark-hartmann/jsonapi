package jsonapi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicPaginator_Params(t *testing.T) {
	page := BasicPaginator{}
	assert.Equal(t, []string{"number", "size"}, page.Params())
}

func TestBasicPaginator_Set(t *testing.T) {
	testData := map[string]struct {
		key           string
		value         string
		expectedValue uint
		expectedError error
	}{
		"valid number value": {
			key:           "number",
			value:         "25",
			expectedValue: 25,
			expectedError: nil,
		},
		"valid size value": {
			key:           "size",
			value:         "15",
			expectedValue: 15,
			expectedError: nil,
		},
		"invalid number value": {
			key:           "number",
			value:         "-25",
			expectedValue: 0,
			expectedError: NewErrInvalidPageNumberParameter("-10"),
		},
		"invalid number value #2": {
			key:           "number",
			value:         "abc",
			expectedValue: 0,
			expectedError: NewErrInvalidPageNumberParameter("abc"),
		},
		"invalid size value": {
			key:           "size",
			value:         "-1",
			expectedValue: 0,
			expectedError: NewErrInvalidPageSizeParameter("-1"),
		},
		"invalid size value #2": {
			key:           "size",
			value:         "abc",
			expectedValue: 0,
			expectedError: NewErrInvalidPageSizeParameter("abc"),
		},
	}

	for name, test := range testData {
		t.Run(name, func(t *testing.T) {

			page := BasicPaginator{}
			err := page.Set(test.key, test.value)

			if test.expectedError != nil {
				assert.Error(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)

				if test.key == "number" {
					assert.Equal(t, test.expectedValue, page.Number, test.key)
					assert.Equal(t, uint(0), page.Size, test.key)
				} else {
					assert.Equal(t, test.expectedValue, page.Size, test.key)
					assert.Equal(t, uint(0), page.Number, test.key)
				}
			}
		})
	}
}

func TestBasicPaginator_Encode(t *testing.T) {
	testData := []struct {
		paginator Paginator
		expected  string
	}{
		{
			paginator: &BasicPaginator{Size: 10, Number: 15},
			expected:  "page%5Bnumber%5D=15&page%5Bsize%5D=10",
		},
		{
			paginator: &BasicPaginator{Number: 15},
			expected:  "page%5Bnumber%5D=15",
		},
		{
			paginator: &BasicPaginator{Size: 15},
			expected:  "page%5Bsize%5D=15",
		},
	}

	for _, test := range testData {
		assert.Equal(t, test.expected, test.paginator.Encode())
	}
}
