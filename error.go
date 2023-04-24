package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var ErrInvalidPayload = errors.New("jsonapi: invalid document payload")

// UnknownTypeError is returned if a type is not known to the schema.
type UnknownTypeError struct {
	Type   string
	inPath bool
}

func (e *UnknownTypeError) Error() string {
	return fmt.Sprintf("jsonapi: resource type %q does not exist", e.Type)
}

// InPath is true if the url pointed to a type that does not exist inside the schema.
func (e *UnknownTypeError) InPath() bool {
	return e.inPath
}

// UnknownFieldError is returned if an attribute or relationship is not known for a given
// resource type.
type UnknownFieldError struct {
	// Type is the name of the resource type the field was not found in.
	Type  string
	Field string

	inPath bool

	// asRel is true if the field is a relationship.
	asRel bool
	relPath
}

func (e *UnknownFieldError) Error() string {
	return fmt.Sprintf("jsonapi: field %q does not exist in resource type %q",
		e.Field, e.Type)
}

// IsAttr returns true if the error was caused by an unknown attribute. Otherwise, it was
// an unknown relationship.
func (e *UnknownFieldError) IsAttr() bool {
	return !e.asRel
}

// InPath is true if the url pointed to a type that does not exist inside the schema.
func (e *UnknownFieldError) InPath() bool {
	return e.inPath
}

// InvalidFieldError is returned if an attribute or relationship is invalid.
type InvalidFieldError struct {
	Type  string
	Field string

	asRel     bool
	wantToOne bool
	isToOne   bool
	relPath
}

func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("jsonapi: field %q of type %q is invalid", e.Field, e.Type)
}

// IsAttr returns true if the error was caused by an invalid attribute. Otherwise, it
// is an invalid relationship.
func (e *InvalidFieldError) IsAttr() bool {
	return !e.asRel
}

// IsInvalidRelType returns true if the error was caused by an invalid relationship
// type, e.g. expected to-one but got to-many.
func (e *InvalidFieldError) IsInvalidRelType() bool {
	return e.isToOne != e.wantToOne
}

// InvalidFieldValueError is returned if a value is to be assigned to an attribute or
// relationship, but the field types do not match.
type InvalidFieldValueError struct {
	Type      string
	Field     string
	FieldType string
	Value     string

	asRel bool

	err error
}

func (e *InvalidFieldValueError) Error() string {
	msg := fmt.Sprintf("jsonapi: invalid value %q for field %q", e.Value, e.Field)
	if e.err != nil {
		msg += ": " + e.err.Error()
	}

	return msg
}

func (e *InvalidFieldValueError) Unwrap() error {
	return e.err
}

// IsAttr returns true if the error was caused by an invalid attribute value. false means
// it was an invalid relationship value.
func (e *InvalidFieldValueError) IsAttr() bool {
	return !e.asRel
}

// IllegalParameterError is returned when a query parameter is used in an illegal
// context. That is, if a collection parameter is used for a single resource or
// if a parameter is not supported.
type IllegalParameterError struct {
	Param      string
	isResource bool
}

func (e *IllegalParameterError) Error() string {
	return fmt.Sprintf("jsonapi: illegal query parameter %q", e.Param)
}

func (e *IllegalParameterError) Source() (string, bool) {
	return e.Param, false
}

// IsResource returns true if the error was caused by using collection parameters (e.g.
// sort or filter) on a single resource endpoint.
func (e *IllegalParameterError) IsResource() bool {
	return e.isResource
}

// ConflictingValueError is returned when two values are mutually exclusive, e.g. if the
// same sort field is used for ascending and descending order.
type ConflictingValueError struct {
	param         string
	value         string
	conflictValue string
}

func (e *ConflictingValueError) Error() string {
	return fmt.Sprintf("jsonapi: conflicting parameter values: %q, %q",
		e.value, e.conflictValue)
}

func (e *ConflictingValueError) Source() (string, bool) {
	return e.param, false
}

// Values returns the conflicting parameter values. If the error was not caused
// by conflicting parameter values, both strings are empty.
func (e *ConflictingValueError) Values() (string, string) {
	return e.value, e.conflictValue
}

// An Error represents an error object from the JSON:API specification.
type Error struct {
	ID     string                 `json:"id"`
	Code   string                 `json:"code"`
	Status string                 `json:"status"`
	Title  string                 `json:"title"`
	Detail string                 `json:"detail"`
	Links  map[string]Link        `json:"links"`
	Source map[string]interface{} `json:"source"`
	Meta   Meta                   `json:"meta"`
}

// NewError returns an empty Error object.
func NewError() Error {
	err := Error{
		Links:  map[string]Link{},
		Source: map[string]interface{}{},
		Meta:   Meta{},
	}

	return err
}

// Error returns the string representation of the error.
//
// If the error does note contain a valid error status code, it returns an empty
// string.
func (e Error) Error() string {
	statusCode, _ := strconv.Atoi(e.Status)
	fullName := http.StatusText(statusCode)

	if fullName != "" && e.Status != "" {
		switch {
		case e.Detail != "":
			return fmt.Sprintf("%s %s: %s", e.Status, fullName, e.Detail)
		case e.Title != "":
			return fmt.Sprintf("%s %s: %s", e.Status, fullName, e.Title)
		default:
			return fmt.Sprintf("%s %s", e.Status, fullName)
		}
	}

	if e.Detail != "" {
		return e.Detail
	}

	return e.Title
}

// MarshalJSON returns a JSON representation of the error according to the
// JSON:API specification.
func (e Error) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}

	if e.ID != "" {
		m["id"] = e.ID
	}

	if e.Code != "" {
		m["code"] = e.Code
	}

	if e.Status != "" {
		m["status"] = e.Status
	}

	if e.Title != "" {
		m["title"] = e.Title
	}

	if e.Detail != "" {
		m["detail"] = e.Detail
	}

	if len(e.Links) > 0 {
		m["links"] = e.Links
	}

	if len(e.Source) > 0 {
		m["source"] = e.Source
	}

	if len(e.Meta) > 0 {
		m["meta"] = e.Meta
	}

	return json.Marshal(m)
}

// NewErrBadRequest (400) returns the corresponding error.
func NewErrBadRequest(title, detail string) Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusBadRequest)
	e.Title = title
	e.Detail = detail

	return e
}

// NewErrUnauthorized (401) returns the corresponding error.
func NewErrUnauthorized() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusUnauthorized)
	e.Title = "Unauthorized"
	e.Detail = "Authentication is required to perform this request."

	return e
}

// NewErrForbidden (403) returns the corresponding error.
func NewErrForbidden() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusForbidden)
	e.Title = "Forbidden"
	e.Detail = "Permission is required to perform this request."

	return e
}

// NewErrNotFound (404) returns the corresponding error.
func NewErrNotFound() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusNotFound)
	e.Title = "Not found"
	e.Detail = "The URI does not exist."

	return e
}

// NewErrPayloadTooLarge (413) returns the corresponding error.
func NewErrPayloadTooLarge() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusRequestEntityTooLarge)
	e.Title = "Payload too large"
	e.Detail = "That's what she said."

	return e
}

// NewErrRequestURITooLong (414) returns the corresponding error.
func NewErrRequestURITooLong() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusRequestURITooLong)
	e.Title = "URI too long"

	return e
}

// NewErrUnsupportedMediaType (415) returns the corresponding error.
func NewErrUnsupportedMediaType() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusUnsupportedMediaType)
	e.Title = "Unsupported media type"

	return e
}

// NewErrTooManyRequests (429) returns the corresponding error.
func NewErrTooManyRequests() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusTooManyRequests)
	e.Title = "Too many requests"

	return e
}

// NewErrRequestHeaderFieldsTooLarge (431) returns the corresponding error.
func NewErrRequestHeaderFieldsTooLarge() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusRequestHeaderFieldsTooLarge)
	e.Title = "Header fields too large"

	return e
}

// NewErrInternalServerError (500) returns the corresponding error.
func NewErrInternalServerError() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusInternalServerError)
	e.Title = "Internal server error"

	return e
}

// NewErrServiceUnavailable (503) returns the corresponding error.
func NewErrServiceUnavailable() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusServiceUnavailable)
	e.Title = "Service unavailable"

	return e
}

// NewErrNotImplemented (503) returns the corresponding error.
func NewErrNotImplemented() Error {
	e := NewError()

	e.Status = strconv.Itoa(http.StatusNotImplemented)
	e.Title = "Not Implemented"

	return e
}

type relPath string

// RelPath returns the relationship path that caused this error. An empty string is
// returned if the error was not caused by an invalid relationship path.
func (e relPath) RelPath() string {
	return string(e)
}

type markedError struct {
	err, mark error
}

func (e *markedError) Error() string {
	return e.err.Error()
}

func (e *markedError) Unwrap() error {
	return e.err
}

func (e *markedError) Is(target error) bool {
	if target == e.mark {
		return true
	}

	return errors.Is(e.err, target)
}

// srcError decorates an existing error with a source.
type srcError struct {
	ptr bool
	src string
	error
}

func (e *srcError) Unwrap() error {
	return e.error
}

// Source returns the error source and a boolean that is true if the source is a
// json pointer.
func (e *srcError) Source() (string, bool) {
	if e.ptr {
		// todo: panic if !isPtr?
		if src, isPtr, ok := errSrc(e.error); ok && isPtr {
			return e.src + src, isPtr
		}

		return e.src, true
	}

	return e.src, false
}

// pathError marks an existing error as caused by an invalid url path.
type pathError struct {
	error
}

func (e *pathError) Unwrap() error {
	return e.error
}

func (e *pathError) InPath() bool {
	return true
}

// errSrc returns the source of the error if applicable.
func errSrc(err error) (src string, isPtr bool, ok bool) {
	var se interface{ Source() (string, bool) }
	if ok = errors.As(err, &se); ok {
		src, isPtr = se.Source()
	}

	return
}

// payloadErr marks any error with ErrInvalidPayload.
func payloadErr(err error) error {
	return &markedError{
		err:  err,
		mark: ErrInvalidPayload,
	}
}

type pathErr interface {
	InPath() bool
}

var _ pathErr = (*pathError)(nil)
var _ pathErr = (*UnknownTypeError)(nil)
var _ pathErr = (*UnknownFieldError)(nil)
