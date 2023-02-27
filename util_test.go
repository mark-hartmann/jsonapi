package jsonapi_test

import (
	"strings"
	"time"
)

func makeOneLineNoSpaces(str string) string {
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "\n", "")

	return strings.ReplaceAll(str, " ", "")
}

func ptr(v interface{}) interface{} {
	switch c := v.(type) {
	// String
	case string:
		return &c
	// Integers
	case int:
		return &c
	case int8:
		return &c
	case int16:
		return &c
	case int32:
		return &c
	case int64:
		return &c
	case uint:
		return &c
	case uint8:
		return &c
	case uint16:
		return &c
	case uint32:
		return &c
	case uint64:
		return &c
	// Bool
	case bool:
		return &c
	// time.Time
	case time.Time:
		return &c
	// []byte
	case []byte:
		return &c
	default:
		return nil
	}
}
