package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Attribute types are the possible types for attributes.
//
// Those constants are numbers that represent the types. Each type has a string
// representation which should be used instead of the numbers when storing
// that information. The numbers can change between any version of this library,
// even if it potentially can break existing code.
//
// The names are as follow:
//   - string
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//   - time (Go type is time.Time)
//   - []byte
//
// An asterisk is present as a prefix if the type is nullable (like *string), brackets
// if it is an array (e.g. []string). Nullable arrays combine asterisk and
// brackets (e.g. *[]string). Byte arrays are represented by []uint8, which are output
// as a base64 encoded string if the attribute type is equal to AttrTypeBytes.
//
// Developers are encouraged to use the constants, the Type struct, and other
// tools to handle attribute types instead of dealing with strings.
const (
	AttrTypeInvalid = iota
	AttrTypeString
	AttrTypeInt
	AttrTypeInt8
	AttrTypeInt16
	AttrTypeInt32
	AttrTypeInt64
	AttrTypeUint
	AttrTypeUint8
	AttrTypeUint16
	AttrTypeUint32
	AttrTypeUint64
	AttrTypeFloat32
	AttrTypeFloat64
	AttrTypeBool
	AttrTypeTime
	AttrTypeBytes
	AttrTypeOther
)

// uint8Array is used to marshal *[]uint8 or []byte as literal array instead of
// a base64 encoded string value.
type uint8Array struct {
	Data     *[]uint8
	Nullable bool
}

func (u uint8Array) MarshalJSON() ([]byte, error) {
	var result string
	if u.Data == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", *u.Data)), ",")
	}

	return []byte(result), nil
}

// A Type stores all the necessary information about a type as represented in
// the JSON:API specification.
//
// NewFunc stores a function that returns a new Resource of the type defined by
// the object with all the fields and the ID set to their zero values. Users may
// call the New method which returns the result of NewFunc if it is non-nil,
// otherwise it returns a SoftResource based on the type.
//
// New makes sure NewFunc is not nil and then calls it, but does not use any
// kind of locking in the process. Therefore, it is unsafe to set NewFunc and
// call New concurrently.
type Type struct {
	Name    string
	Attrs   map[string]Attr
	Rels    map[string]Rel
	NewFunc func() Resource
}

// AddAttr adds an attributes to the type.
func (t *Type) AddAttr(attr Attr) error {
	// Validation
	if attr.Name == "" {
		return fmt.Errorf("jsonapi: attribute name is empty")
	}

	if attr.Unmarshaler == nil && GetAttrTypeString(attr.Type, attr.Array, attr.Nullable) == "" {
		return fmt.Errorf("jsonapi: attribute type is invalid")
	}

	// Make sure the name isn't already used
	for i := range t.Attrs {
		if t.Attrs[i].Name == attr.Name {
			return fmt.Errorf("jsonapi: attribute name %q is already used", attr.Name)
		}
	}

	if t.Attrs == nil {
		t.Attrs = map[string]Attr{}
	}

	t.Attrs[attr.Name] = attr

	return nil
}

// RemoveAttr removes an attribute from the type if it exists.
func (t *Type) RemoveAttr(attr string) {
	for i := range t.Attrs {
		if t.Attrs[i].Name == attr {
			delete(t.Attrs, attr)
		}
	}
}

// AddRel adds a relationship to the type.
func (t *Type) AddRel(rel Rel) error {
	// Validation
	if rel.FromName == "" {
		return fmt.Errorf("jsonapi: relationship name is empty")
	}

	if rel.ToType == "" {
		return fmt.Errorf("jsonapi: relationship type is empty")
	}

	// Make sure the name isn't already used
	for i := range t.Rels {
		if t.Rels[i].FromName == rel.FromName {
			return fmt.Errorf("jsonapi: relationship name %q is already used", rel.FromName)
		}
	}

	if t.Rels == nil {
		t.Rels = map[string]Rel{}
	}

	t.Rels[rel.FromName] = rel

	return nil
}

// RemoveRel removes a relationship from the type if it exists.
func (t *Type) RemoveRel(rel string) {
	for i := range t.Rels {
		if t.Rels[i].FromName == rel {
			delete(t.Rels, rel)
		}
	}
}

// Fields returns a list of the names of all the fields (attributes and
// relationships) in the type.
func (t *Type) Fields() []string {
	fields := make([]string, 0, len(t.Attrs)+len(t.Rels))
	for i := range t.Attrs {
		fields = append(fields, t.Attrs[i].Name)
	}

	for i := range t.Rels {
		fields = append(fields, t.Rels[i].FromName)
	}

	sort.Strings(fields)

	return fields
}

// New calls the NewFunc field and returns the result Resource object.
//
// If NewFunc is nil, it returns a *SoftResource with its Type field set to the
// value of the receiver.
func (t *Type) New() Resource {
	if t.NewFunc != nil {
		return t.NewFunc()
	}

	return &SoftResource{Type: t}
}

// Equal returns true if both types have the same name, attributes,
// relationships. NewFunc is ignored.
func (t Type) Equal(typ Type) bool {
	t.NewFunc = nil
	typ.NewFunc = nil

	return reflect.DeepEqual(t, typ)
}

// Copy deeply copies the receiver and returns the result.
func (t Type) Copy() Type {
	ctyp := Type{
		Name:  t.Name,
		Attrs: map[string]Attr{},
		Rels:  map[string]Rel{},
	}

	for name, attr := range t.Attrs {
		ctyp.Attrs[name] = attr
	}

	for name, rel := range t.Rels {
		ctyp.Rels[name] = rel
	}

	ctyp.NewFunc = t.NewFunc

	return ctyp
}

// TypeUnmarshaler is a custom type unmarshaler that is used instead of Attr.UnmarshalToType for
// types that are not supported by default.
type TypeUnmarshaler interface {
	GetZeroValue(array, nullable bool) interface{}
	UnmarshalToType(data []byte, array, nullable bool) (interface{}, error)
}

// TypeNameExposer allows a custom unmarshaler to specify the real type name or referenceable alias
// in the error metadata if the value cannot be unmarshalled correctly.
type TypeNameExposer interface {
	PublicTypeName() string
}

// Attr represents a resource attribute.
type Attr struct {
	Name        string
	Type        int
	Nullable    bool
	Array       bool
	Unmarshaler TypeUnmarshaler
}

// UnmarshalToType unmarshalls the data into a value of the type represented by
// the attribute and returns it.
func (a Attr) UnmarshalToType(data []byte) (interface{}, error) {
	if data == nil || (!a.Nullable && string(data) == "null") {
		return nil, fmt.Errorf("%s is not nullable", a.Name)
	}

	if a.Nullable && string(data) == "null" {
		return GetZeroValue(a.Type, a.Array, a.Nullable), nil
	}

	var (
		v   interface{}
		err error
	)

	switch a.Type {
	case AttrTypeString:
		if a.Array {
			var sa []string
			err = json.Unmarshal(data, &sa)

			if a.Nullable {
				v = &sa
			} else {
				v = sa
			}
		} else if string(data) != "null" {
			var s string
			err = json.Unmarshal(data, &s)

			if a.Nullable {
				v = &s
			} else {
				v = s
			}
		}
	case AttrTypeInt:
		if a.Array {
			var ia []int
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if a.Nullable {
				n := v.(int)
				v = &n
			} else {
				v = v.(int)
			}
		}
	case AttrTypeInt8:
		if a.Array {
			var ia []int8
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if a.Nullable {
				n := int8(v.(int))
				v = &n
			} else {
				v = int8(v.(int))
			}
		}
	case AttrTypeInt16:
		if a.Array {
			var ia []int16
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if a.Nullable {
				n := int16(v.(int))
				v = &n
			} else {
				v = int16(v.(int))
			}
		}
	case AttrTypeInt32:
		if a.Array {
			var ia []int32
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if a.Nullable {
				n := int32(v.(int))
				v = &n
			} else {
				v = int32(v.(int))
			}
		}
	case AttrTypeInt64:
		if a.Array {
			var ia []int64
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.Atoi(string(data))

			if a.Nullable {
				n := int64(v.(int))
				v = &n
			} else {
				v = int64(v.(int))
			}
		}
	case AttrTypeUint:
		if a.Array {
			var ia []uint
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 64)

			if a.Nullable {
				n := uint(v.(uint64))
				v = &n
			} else {
				v = uint(v.(uint64))
			}
		}
	case AttrTypeUint8:
		if a.Array {
			var ia []uint8
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 8)

			if a.Nullable {
				n := uint8(v.(uint64))
				v = &n
			} else {
				v = uint8(v.(uint64))
			}
		}
	case AttrTypeUint16:
		if a.Array {
			var ia []uint16
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 16)

			if a.Nullable {
				n := uint16(v.(uint64))
				v = &n
			} else {
				v = uint16(v.(uint64))
			}
		}
	case AttrTypeUint32:
		if a.Array {
			var ia []uint32
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 32)

			if a.Nullable {
				n := uint32(v.(uint64))
				v = &n
			} else {
				v = uint32(v.(uint64))
			}
		}
	case AttrTypeUint64:
		if a.Array {
			var ia []uint64
			err = json.Unmarshal(data, &ia)

			if a.Nullable {
				v = &ia
			} else {
				v = ia
			}
		} else {
			v, err = strconv.ParseUint(string(data), 10, 64)

			if a.Nullable {
				n := v.(uint64)
				v = &n
			} else {
				v = v.(uint64)
			}
		}
	case AttrTypeFloat32:
		if a.Array {
			var fa []float32
			err = json.Unmarshal(data, &fa)

			if a.Nullable {
				v = &fa
			} else {
				v = fa
			}
		} else {
			var f64 float64
			f64, err = strconv.ParseFloat(string(data), 32)
			if a.Nullable {
				n := float32(f64)
				v = &n
			} else {
				v = float32(f64)
			}
		}
	case AttrTypeFloat64:
		if a.Array {
			var fa []float64
			err = json.Unmarshal(data, &fa)

			if a.Nullable {
				v = &fa
			} else {
				v = fa
			}
		} else {
			var f64 float64
			f64, err = strconv.ParseFloat(string(data), 64)
			if a.Nullable {
				n := f64
				v = &n
			} else {
				v = f64
			}
		}
	case AttrTypeBool:
		if a.Array {
			var ba []bool
			err = json.Unmarshal(data, &ba)

			if a.Nullable {
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

			if a.Nullable {
				v = &b
			} else {
				v = b
			}
		}
	case AttrTypeTime:
		if a.Array {
			var ta []time.Time
			err = json.Unmarshal(data, &ta)

			if a.Nullable {
				v = &ta
			} else {
				v = ta
			}
		} else {
			var t time.Time
			err = json.Unmarshal(data, &t)

			if a.Nullable {
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

		if a.Nullable {
			v = &s
		} else {
			v = s
		}
	default:
		err = errors.New("attribute is of invalid or unknown type")
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

// Rel represents a resource relationship.
type Rel struct {
	FromType string
	FromName string
	ToOne    bool
	ToType   string
	ToName   string
	FromOne  bool
}

// Invert returns the inverse relationship of r.
func (r *Rel) Invert() Rel {
	return Rel{
		FromType: r.ToType,
		FromName: r.ToName,
		ToOne:    r.FromOne,
		ToType:   r.FromType,
		ToName:   r.FromName,
		FromOne:  r.ToOne,
	}
}

// Normalize inverts the relationship if necessary in order to have it in the
// right direction and returns the result.
//
// This is the form stored in Schema.Rels.
func (r *Rel) Normalize() Rel {
	from := r.FromType + r.FromName
	to := r.ToType + r.ToName

	if from < to || r.ToName == "" {
		return *r
	}

	return r.Invert()
}

// String builds and returns the name of the receiving Rel.
//
// r.Normalize is always called.
func (r Rel) String() string {
	r = r.Normalize()

	id := r.FromType + "_" + r.FromName
	if r.ToName != "" {
		id += "_" + r.ToType + "_" + r.ToName
	}

	return id
}

// GetAttrType returns the attribute type as int (see constants) and whether
// the type is an array and/or nullable (ptr).
func GetAttrType(t string) (typ int, array bool, nullable bool) {
	bi := strings.Index(t, "[]")
	array = bi == 0 || bi == 1
	nullable = strings.HasPrefix(t, "*")

	switch {
	case nullable && array:
		t = t[3:]
	case array:
		t = t[2:]
	case nullable:
		t = t[1:]
	}

	switch t {
	case "string":
		return AttrTypeString, array, nullable
	case "int":
		return AttrTypeInt, array, nullable
	case "int8":
		return AttrTypeInt8, array, nullable
	case "int16":
		return AttrTypeInt16, array, nullable
	case "int32":
		return AttrTypeInt32, array, nullable
	case "int64":
		return AttrTypeInt64, array, nullable
	case "uint":
		return AttrTypeUint, array, nullable
	case "uint8", "byte":
		return AttrTypeUint8, array, nullable
	case "uint16":
		return AttrTypeUint16, array, nullable
	case "uint32":
		return AttrTypeUint32, array, nullable
	case "uint64":
		return AttrTypeUint64, array, nullable
	case "float32":
		return AttrTypeFloat32, array, nullable
	case "float64":
		return AttrTypeFloat64, array, nullable
	case "bool":
		return AttrTypeBool, array, nullable
	case "time.Time":
		return AttrTypeTime, array, nullable
	default:
		return AttrTypeInvalid, false, false
	}
}

// GetAttrTypeString returns the name of the attribute type specified by t (see
// constants), array and nullable.
func GetAttrTypeString(t int, array, nullable bool) string {
	str := ""

	switch t {
	case AttrTypeString:
		str = "string"
	case AttrTypeInt:
		str = "int"
	case AttrTypeInt8:
		str = "int8"
	case AttrTypeInt16:
		str = "int16"
	case AttrTypeInt32:
		str = "int32"
	case AttrTypeInt64:
		str = "int64"
	case AttrTypeUint:
		str = "uint"
	case AttrTypeUint8:
		str = "uint8"
	case AttrTypeBytes:
		array = true
		str = "uint8"
	case AttrTypeUint16:
		str = "uint16"
	case AttrTypeUint32:
		str = "uint32"
	case AttrTypeUint64:
		str = "uint64"
	case AttrTypeFloat32:
		str = "float32"
	case AttrTypeFloat64:
		str = "float64"
	case AttrTypeBool:
		str = "bool"
	case AttrTypeTime:
		str = "time.Time"
	default:
		str = ""
	}

	if array {
		str = "[]" + str
	}

	if nullable {
		return "*" + str
	}

	return str
}

// GetZeroValue returns the zero value of the attribute type represented by the
// specified int (see constants).
//
// If nullable is true, the returned value is a nil pointer. if nullable and array
// are true, a null pointer to a slice is returned.
func GetZeroValue(t int, array, nullable bool) interface{} {
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

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	}

	return false
}
