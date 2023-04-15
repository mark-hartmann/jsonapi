package jsonapi

import (
	"fmt"
	"net/http"
)

// NewRequest builds a *Request based on a *http.Request and validated by a *Schema.
func NewRequest(r *http.Request, schema *Schema) (*Request, error) {
	su, err := NewSimpleURL(r.URL)
	if err != nil {
		return nil, fmt.Errorf("jsonapi: failed to create jsonapi.SimpleURL: %w", err)
	}

	url, err := NewURL(schema, su)
	if err != nil {
		return nil, fmt.Errorf("jsonapi: failed to create jsonapi.URL: %w", err)
	}

	var doc *Document

	if r.Method == http.MethodPost || r.Method == http.MethodPatch {
		doc, err = UnmarshalDocument(r.Body, schema)
		if err != nil {
			return nil, fmt.Errorf("jsonapi: failed to unmarshal request body: %w", err)
		}
	}

	req := &Request{
		Method: r.Method,
		URL:    url,
		Doc:    doc,
	}

	return req, nil
}

// A Request represents a JSON:API request.
type Request struct {
	Method string
	URL    *URL
	Doc    *Document
}
