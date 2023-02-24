package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var registry = typeRegistry{
	names:       map[int]string{},
	namesR:      map[string]int{},
	values:      map[int]ZeroValueFunc{},
	unmarshaler: map[int]TypeUnmarshalerFunc{},
	nameFunc: func(name string, array, nullable bool) string {
		if array && nullable {
			return name + "[] (nullable)"
		} else if nullable {
			return name + " (nullable)"
		} else if array {
			return name + "[]"
		}
		return name
	},
}

func init() {
	// Int
	RegisterAttrType(AttrTypeInt, "int", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeInt8, "int8", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeInt16, "int16", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeInt32, "int32", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeInt64, "int64", basicZeroValueFunc, basicUnmarshalerFunc)
	// Uint
	RegisterAttrType(AttrTypeUint, "uint", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeUint8, "uint8", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeUint16, "uint16", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeUint32, "uint32", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeUint64, "uint64", basicZeroValueFunc, basicUnmarshalerFunc)
	// Float
	RegisterAttrType(AttrTypeFloat32, "float32", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeFloat64, "float64", basicZeroValueFunc, basicUnmarshalerFunc)
	// Misc
	RegisterAttrType(AttrTypeString, "string", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeBool, "bool", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeTime, "time", basicZeroValueFunc, basicUnmarshalerFunc)
	RegisterAttrType(AttrTypeBytes, "bytes", basicZeroValueFunc, basicUnmarshalerFunc)
}

// NameFunc receives the name of an attribute type and can replace or extend it to add context.
type NameFunc func(name string, array, nullable bool) string

// ZeroValueFunc returns the null value of the attribute type for any possible combination of the
// nullable and array parameters.
type ZeroValueFunc func(typ int, array, nullable bool) interface{}

// TypeUnmarshalerFunc will unmarshal attribute payload to an appropriate golang type.
type TypeUnmarshalerFunc func(data []byte, attr Attr) (interface{}, error)

type typeRegistry struct {
	names       map[int]string
	namesR      map[string]int
	values      map[int]ZeroValueFunc
	unmarshaler map[int]TypeUnmarshalerFunc
	nameFunc    NameFunc
}

// RegisterAttrType registers a new attribute type or overrides a previously registered one.
// Calling this function without a valid name or ZeroValueFunc / TypeUnmarshalerFunc will panic.
func RegisterAttrType(typ int, name string, zeroValueFunc ZeroValueFunc,
	unmarshalerFn TypeUnmarshalerFunc) {
	if typ == AttrTypeInvalid || name == "" || zeroValueFunc == nil || unmarshalerFn == nil {
		panic(fmt.Sprintf("jsonapi: failed to register attribute type %q", typ))
	}

	registry.names[typ] = name
	registry.namesR[name] = typ

	registry.values[typ] = zeroValueFunc
	registry.unmarshaler[typ] = unmarshalerFn
}

// GetZeroValue returns the zero value of the attribute type represented by the
// specified int (see constants and RegisterAttrType).
//
// If nullable is true, the returned value is a nil pointer. if nullable and array
// are true, a null pointer to a slice is returned. The zero value refers to the
// JSON attributes and not to the go types, so an array is not nil, but empty.
func GetZeroValue(typ int, array, nullable bool) (interface{}, error) {
	fn, ok := registry.values[typ]
	if !ok {
		return nil, fmt.Errorf("jsonapi: unregistered attribute type %q", typ)
	}

	return fn(typ, array, nullable), nil
}

// UnmarshalToType unmarshalls the data into a value of the type represented by the attribute.
func UnmarshalToType(data []byte, attr Attr) (interface{}, error) {
	fn, ok := registry.unmarshaler[attr.Type]
	if !ok {
		return nil, fmt.Errorf("jsonapi: unregistered attribute type %q", attr.Type)
	}

	return fn(data, attr)
}

// GetAttrTypeName returns the public name for the attribute type. If set, the name is processed
// by the NameFunc before it is returned.
func GetAttrTypeName(typ int, array, nullable bool) (string, error) {
	name, ok := registry.names[typ]
	if !ok {
		return "", fmt.Errorf("jsonapi: unregistered attribute type %q", typ)
	}

	if registry.nameFunc != nil {
		return registry.nameFunc(name, array, nullable), nil
	}

	return name, nil
}

// SetAttrTypeNameFunc overwrites the registry's NameFunc.
func SetAttrTypeNameFunc(fn NameFunc) {
	registry.nameFunc = fn
}

func attrTypeRegistered(typ int) bool {
	_, ok := registry.names[typ]
	return ok
}

// basicUnmarshalerFunc is the default TypeUnmarshalerFunc for all attribute types that are
// supported by jsonapi out of the box (see constants).
func basicUnmarshalerFunc(data []byte, attr Attr) (interface{}, error) {
	if data == nil || (!attr.Nullable && string(data) == "null") {
		return nil, fmt.Errorf("%s is not nullable", attr.Name)
	}

	if attr.Nullable && string(data) == "null" {
		return GetZeroValue(attr.Type, attr.Array, attr.Nullable)
	}

	var (
		v   interface{}
		err error
	)

	switch attr.Type {
	case AttrTypeString:
		if attr.Array {
			var sa []string
			err = json.Unmarshal(data, &sa)

			if attr.Nullable {
				v = &sa
			} else {
				v = sa
			}
		} else if string(data) != "null" {
			var s string
			err = json.Unmarshal(data, &s)

			if attr.Nullable {
				v = &s
			} else {
				v = s
			}
		}
	case AttrTypeInt:
		if attr.Array {
			var ia []int
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if attr.Nullable {
				n := v.(int)
				v = &n
			} else {
				v = v.(int)
			}
		}
	case AttrTypeInt8:
		if attr.Array {
			var ia []int8
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if attr.Nullable {
				n := int8(v.(int))
				v = &n
			} else {
				v = int8(v.(int))
			}
		}
	case AttrTypeInt16:
		if attr.Array {
			var ia []int16
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if attr.Nullable {
				n := int16(v.(int))
				v = &n
			} else {
				v = int16(v.(int))
			}
		}
	case AttrTypeInt32:
		if attr.Array {
			var ia []int32
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if attr.Nullable {
				n := int32(v.(int))
				v = &n
			} else {
				v = int32(v.(int))
			}
		}
	case AttrTypeInt64:
		if attr.Array {
			var ia []int64
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if attr.Nullable {
				n := int64(v.(int))
				v = &n
			} else {
				v = int64(v.(int))
			}
		}
	case AttrTypeUint:
		if attr.Array {
			var ia []uint
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 64)

			if attr.Nullable {
				n := uint(v.(uint64))
				v = &n
			} else {
				v = uint(v.(uint64))
			}
		}
	case AttrTypeUint8:
		if attr.Array {
			var ia []uint8
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 8)

			if attr.Nullable {
				n := uint8(v.(uint64))
				v = &n
			} else {
				v = uint8(v.(uint64))
			}
		}
	case AttrTypeUint16:
		if attr.Array {
			var ia []uint16
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 16)

			if attr.Nullable {
				n := uint16(v.(uint64))
				v = &n
			} else {
				v = uint16(v.(uint64))
			}
		}
	case AttrTypeUint32:
		if attr.Array {
			var ia []uint32
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 32)

			if attr.Nullable {
				n := uint32(v.(uint64))
				v = &n
			} else {
				v = uint32(v.(uint64))
			}
		}
	case AttrTypeUint64:
		if attr.Array {
			var ia []uint64
			err = json.Unmarshal(data, &ia)

			if attr.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 64)

			if attr.Nullable {
				n := v.(uint64)
				v = &n
			} else {
				v = v.(uint64)
			}
		}
	case AttrTypeFloat32:
		if attr.Array {
			var fa []float32
			err = json.Unmarshal(data, &fa)

			if attr.Nullable {
				v = &fa
			} else {
				v = fa
			}
		} else {
			var f64 float64
			f64, err = strconv.ParseFloat(string(data), 32)
			if attr.Nullable {
				n := float32(f64)
				v = &n
			} else {
				v = float32(f64)
			}
		}
	case AttrTypeFloat64:
		if attr.Array {
			var fa []float64
			err = json.Unmarshal(data, &fa)

			if attr.Nullable {
				v = &fa
			} else {
				v = fa
			}
		} else {
			var f64 float64
			f64, err = strconv.ParseFloat(string(data), 64)
			if attr.Nullable {
				n := f64
				v = &n
			} else {
				v = f64
			}
		}
	case AttrTypeBool:
		if attr.Array {
			var ba []bool
			err = json.Unmarshal(data, &ba)

			if attr.Nullable {
				v = &ba
			} else {
				v = ba
			}
		} else {
			var b bool
			if string(data) == "true" {
				b = true
			} else if string(data) != "false" {
				err = errors.New("boolean is not true or false")
			}

			if attr.Nullable {
				v = &b
			} else {
				v = b
			}
		}
	case AttrTypeTime:
		if attr.Array {
			var ta []time.Time
			err = json.Unmarshal(data, &ta)

			if attr.Nullable {
				v = &ta
			} else {
				v = ta
			}
		} else {
			var t time.Time
			err = json.Unmarshal(data, &t)

			if attr.Nullable {
				v = &t
			} else {
				v = t
			}
		}
	case AttrTypeBytes:
		s := make([]byte, len(data))
		err := json.Unmarshal(data, &s)

		if err != nil {
			panic(err)
		}

		if attr.Nullable {
			v = &s
		} else {
			v = s
		}
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

// basicZeroValueFunc is the default ZeroValueFunc for all attribute types that are supported by
// jsonapi out of the box (see constants).
func basicZeroValueFunc(t int, array, nullable bool) interface{} {
	switch t {
	case AttrTypeString:
		switch {
		case nullable && array:
			return (*[]string)(nil)
		case array:
			return []string{}
		case nullable:
			return (*string)(nil)
		}

		return ""
	case AttrTypeInt:
		switch {
		case nullable && array:
			return (*[]int)(nil)
		case array:
			return []int{}
		case nullable:
			return (*int)(nil)
		}

		return 0
	case AttrTypeInt8:
		switch {
		case nullable && array:
			return (*[]int8)(nil)
		case array:
			return []int8{}
		case nullable:
			return (*int8)(nil)
		}

		return int8(0)
	case AttrTypeInt16:
		switch {
		case nullable && array:
			return (*[]int16)(nil)
		case array:
			return []int16{}
		case nullable:
			return (*int16)(nil)
		}

		return int16(0)
	case AttrTypeInt32:
		switch {
		case nullable && array:
			return (*[]int32)(nil)
		case array:
			return []int32{}
		case nullable:
			return (*int32)(nil)
		}

		return int32(0)
	case AttrTypeInt64:
		switch {
		case nullable && array:
			return (*[]int64)(nil)
		case array:
			return []int64{}
		case nullable:
			return (*int64)(nil)
		}

		return int64(0)
	case AttrTypeUint:
		switch {
		case nullable && array:
			return (*[]uint)(nil)
		case array:
			return []uint{}
		case nullable:
			return (*uint)(nil)
		}

		return uint(0)
	case AttrTypeUint8, AttrTypeBytes:
		if t == AttrTypeBytes {
			array = true
		}

		switch {
		case nullable && array:
			return (*[]uint8)(nil)
		case array:
			return []uint8{}
		case nullable:
			return (*uint8)(nil)
		}

		return uint8(0)
	case AttrTypeUint16:
		switch {
		case nullable && array:
			return (*[]uint16)(nil)
		case array:
			return []uint16{}
		case nullable:
			return (*uint16)(nil)
		}

		return uint16(0)
	case AttrTypeUint32:
		switch {
		case nullable && array:
			return (*[]uint32)(nil)
		case array:
			return []uint32{}
		case nullable:
			return (*uint32)(nil)
		}

		return uint32(0)
	case AttrTypeUint64:
		switch {
		case nullable && array:
			return (*[]uint64)(nil)
		case array:
			return []uint64{}
		case nullable:
			return (*uint64)(nil)
		}

		return uint64(0)
	case AttrTypeFloat32:
		switch {
		case nullable && array:
			return (*[]float32)(nil)
		case array:
			return []float32{}
		case nullable:
			return (*float32)(nil)
		}

		return float32(0)
	case AttrTypeFloat64:
		switch {
		case nullable && array:
			return (*[]float64)(nil)
		case array:
			return []float64{}
		case nullable:
			return (*float64)(nil)
		}

		return float64(0)
	case AttrTypeBool:
		switch {
		case nullable && array:
			return (*[]bool)(nil)
		case array:
			return []bool{}
		case nullable:
			return (*bool)(nil)
		}

		return false
	case AttrTypeTime:
		switch {
		case nullable && array:
			return (*[]time.Time)(nil)
		case array:
			return []time.Time{}
		case nullable:
			return (*time.Time)(nil)
		}

		return time.Time{}
	default:
		return nil
	}
}
