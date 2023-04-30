package jsonapi_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	. "github.com/mark-hartmann/jsonapi"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name           string
		err            Error
		expectedString string
	}{
		{
			name: "empty",
			err: func() Error {
				e := NewError()
				return e
			}(),
			expectedString: "",
		}, {
			name: "title",
			err: func() Error {
				e := NewError()
				e.Title = "An error"
				return e
			}(),
			expectedString: "An error",
		}, {
			name: "detail",
			err: func() Error {
				e := NewError()
				e.Detail = "An error occurred."
				return e
			}(),
			expectedString: "An error occurred.",
		}, {
			name: "http status code",
			err: func() Error {
				e := NewError()
				e.Status = strconv.Itoa(http.StatusInternalServerError)
				return e
			}(),
			expectedString: "500 Internal Server Error",
		}, {
			name: "http status code and title",
			err: func() Error {
				e := NewError()
				e.Status = strconv.Itoa(http.StatusInternalServerError)
				e.Title = "Internal server error"
				return e
			}(),
			expectedString: "500 Internal Server Error: Internal server error",
		}, {
			name: "http status code and detail",
			err: func() Error {
				e := NewError()
				e.Status = strconv.Itoa(http.StatusInternalServerError)
				e.Detail = "An internal server error occurred."
				return e
			}(),
			expectedString: "500 Internal Server Error: An internal server error occurred.",
		},
	}

	for _, test := range tests {
		assert.Equal(test.err.Error(), test.expectedString, test.name)
	}
}

func TestErrorConstructors(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name: "NewError",
			err: func() Error {
				e := NewError()
				return e
			}(),
			expected: "",
		}, {
			name: "NewErrBadRequest",
			err: func() Error {
				e := NewErrBadRequest("bad request", "error detail")
				return e
			}(),
			expected: "400 Bad Request: error detail",
		}, {
			name: "NewErrUnauthorized",
			err: func() Error {
				e := NewErrUnauthorized()
				return e
			}(),
			expected: "401 Unauthorized: Authentication is required to perform this request.",
		}, {
			name: "NewErrForbidden",
			err: func() Error {
				e := NewErrForbidden()
				return e
			}(),
			expected: "403 Forbidden: Permission is required to perform this request.",
		}, {
			name: "NewErrNotFound",
			err: func() Error {
				e := NewErrNotFound()
				return e
			}(),
			expected: "404 Not Found: The URI does not exist.",
		}, {
			name: "NewErrPayloadTooLarge",
			err: func() Error {
				e := NewErrPayloadTooLarge()
				return e
			}(),
			expected: "413 Request Entity Too Large: That's what she said.",
		}, {
			name: "NewErrRequestURITooLong",
			err: func() Error {
				e := NewErrRequestURITooLong()
				return e
			}(),
			expected: "414 Request URI Too Long: URI too long",
		}, {
			name: "NewErrUnsupportedMediaType",
			err: func() Error {
				e := NewErrUnsupportedMediaType()
				return e
			}(),
			expected: "415 Unsupported Media Type: Unsupported media type",
		}, {
			name: "NewErrTooManyRequests",
			err: func() Error {
				e := NewErrTooManyRequests()
				return e
			}(),
			expected: "429 Too Many Requests: Too many requests",
		}, {
			name: "NewErrRequestHeaderFieldsTooLarge",
			err: func() Error {
				e := NewErrRequestHeaderFieldsTooLarge()
				return e
			}(),
			expected: "431 Request Header Fields Too Large: Header fields too large",
		}, {
			name: "NewErrInternalServerError",
			err: func() Error {
				e := NewErrInternalServerError()
				return e
			}(),
			expected: "500 Internal Server Error: Internal server error",
		}, {
			name: "NewErrServiceUnavailable",
			err: func() Error {
				e := NewErrServiceUnavailable()
				return e
			}(),
			expected: "503 Service Unavailable: Service unavailable",
		}, {
			name: "NewErrNotImplemented",
			err: func() Error {
				e := NewErrNotImplemented()
				return e
			}(),
			expected: "501 Not Implemented: Not Implemented",
		},
	}

	for _, test := range tests {
		assert.Equal(test.expected, test.err.Error(), test.name)
	}
}

func TestErrorMarshalJSON(t *testing.T) {
	assert := assert.New(t)

	jaerr := Error{
		ID:     "c1897530-fdf5-4a42-88fb-1c1c4bd0962f",
		Code:   "Code",
		Status: "Status",
		Title:  "Title",
		Detail: "Detail",
		Links: map[string]Link{
			"link": {
				HRef: "https://example.com",
			},
			"about": {
				HRef: "https://example.com",
				Meta: map[string]interface{}{
					"abc": "def",
				},
			},
		},
		Source: map[string]interface{}{
			"parameter": "param",
			"pointer":   "/data",
		},
		Meta: map[string]interface{}{
			"meta": 123,
		},
	}

	payload, err := json.Marshal(jaerr)
	assert.NoError(err)
	assert.JSONEq(string(payload), `
		{
			"code": "Code",
			"detail": "Detail",
			"id": "c1897530-fdf5-4a42-88fb-1c1c4bd0962f",
			"links": {
				"link": "https://example.com",
				"about": {
					"href": "https://example.com",
					"meta": {
						"abc": "def"					
					}
				}
			},
			"meta": {
				"meta": 123
			},
			"source": {
				"parameter": "param",
				"pointer": "/data"
			},
			"status": "Status",
			"title": "Title"
		}
	`)
}
