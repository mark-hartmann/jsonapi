package jsonapi_test

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	assert := assert.New(t)

	// Schema
	schema := newMockSchema()

	tests := []struct {
		name          string
		method        string
		url           string
		schema        *Schema
		expectedError string
	}{
		{
			name:          "get collection (mock schema)",
			method:        "GET",
			url:           "/mocktypes1",
			schema:        schema,
			expectedError: "",
		}, {
			name:   "bad url",
			method: "GET",
			url:    "/invalid",
			schema: schema,
			expectedError: "" +
				"jsonapi: failed to create jsonapi.URL: " +
				`jsonapi: resource type "invalid" does not exist`,
		},
	}

	for _, test := range tests {
		body := bytes.NewBufferString("")
		req := httptest.NewRequest(test.method, test.url, body)

		doc, err := NewRequest(req, test.schema)

		if test.expectedError == "" {
			assert.NoError(err)
			assert.Equal(test.method, doc.Method, test.name)
		} else {
			assert.EqualError(err, test.expectedError, test.name)
			assert.Nil(doc)
		}
	}

	body := bytes.NewBufferString("")
	req := httptest.NewRequest("GET", "/mocktypes1", body)
	req.URL = nil
	doc, err := NewRequest(req, schema)
	assert.EqualError(err, ""+
		"jsonapi: failed to create jsonapi.SimpleURL: "+
		"jsonapi: pointer to url.URL is nil")
	assert.Nil(doc)
}

func TestNewRequestInvalidBody(t *testing.T) {
	assert := assert.New(t)

	// Schema
	schema := newMockSchema()

	// Nil body
	req := httptest.NewRequest("POST", "/mocktypes1", badReader{})

	doc, err := NewRequest(req, schema)
	assert.EqualError(err, "jsonapi: failed to unmarshal request body: bad reader")
	assert.Nil(doc)

	// Invalid body
	body := bytes.NewBufferString("{invalidjson}")
	req = httptest.NewRequest("POST", "/mocktypes1", body)

	doc, err = NewRequest(req, schema)
	assert.EqualError(err, ""+
		"jsonapi: failed to unmarshal request body: invalid character 'i' looking for "+
		"beginning of object key string",
	)
	assert.Nil(doc)
}

type badReader struct {
}

func (badReader) Read([]byte) (int, error) {
	return 0, errors.New("bad reader")
}
