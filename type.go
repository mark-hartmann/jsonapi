package jsonapi

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
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
	AttrTypeInvalid = (0 ^ iota) * -1
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
)

var memberRegexp = regexp.MustCompile(`^[a-zA-Z0-9](?:[-\w]*[a-zA-Z0-9])?$`)

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
	if !memberRegexp.MatchString(attr.Name) {
		return fmt.Errorf("jsonapi: attribute name does not meet member name requirements")
	}

	// SPEC 5.2.2 - 5.2.3
	// The user is responsible for ensuring that types which are turned into a json object have
	// neither "relationships" nor "links" as field names.
	if attr.Name == "relationships" || attr.Name == "links" || attr.Name == "id" ||
		attr.Name == "type" {
		return fmt.Errorf(`jsonapi: illegal attribute name "%s"`, attr.Name)
	}

	if attr.Type == AttrTypeInvalid {
		return fmt.Errorf("jsonapi: cannot add attribute with type AttrTypeInvalid")
	}

	if !attrTypeRegistered(attr.Type) {
		return fmt.Errorf("jsonapi: attribute type %q is unknown", attr.Type)
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
	if !memberRegexp.MatchString(rel.FromName) {
		return fmt.Errorf("jsonapi: relationship name does not meet member " +
			"name requirements")
	}

	// SPEC 5.2.2 - 5.2.3
	if rel.FromName == "id" || rel.FromName == "type" {
		return fmt.Errorf(`jsonapi: illegal relationship name "%s"`, rel.FromName)
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

// Attr represents a resource attribute.
type Attr struct {
	Name     string
	Type     int
	Nullable bool
	Array    bool
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

// SortRule is a representation of a sorting rule.
//
// SPEC 6.5
type SortRule struct {
	// Path contains, if it is a relationship attribute based sorting rule, the
	// complete relationship path.
	Path []Rel
	// Name is the name of the sort field (attribute)
	Name string
	Desc bool
}

// GetAttrType returns the attribute type as int (see constants) and whether
// the type is an array and/or nullable (ptr).
//
// Deprecated: This function will be removed or unexported, as it doesn't really work with the
// new type functionality.
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
		return AttrTypeInvalid, array, nullable
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
