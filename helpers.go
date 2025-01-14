package jsonapi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Check checks that the given value can be used with this library and returns
// the first error it finds.
//
// It makes sure that the struct has an ID field of type string and that the api
// key of the field tags are properly formatted.
//
// If nil is returned, then the value can be safely used with this library.
func Check(v interface{}) error {
	value := reflect.ValueOf(v)
	kind := value.Kind()

	// Check whether it's a struct
	if kind != reflect.Struct {
		return errors.New("jsonapi: not a struct")
	}

	// Check ID field
	var (
		idField reflect.StructField
		ok      bool
	)

	if idField, ok = value.Type().FieldByName("ID"); !ok {
		return errors.New("jsonapi: struct doesn't have an ID field")
	}

	resType := idField.Tag.Get("api")
	if resType == "" {
		return errors.New("jsonapi: ID field's api tag is empty")
	}

	// Check attributes
	for i := 0; i < value.NumField(); i++ {
		sf := value.Type().Field(i)
		typ := sf.Type

		switch typ.Kind() {
		case reflect.Ptr:
			typ = typ.Elem()
			if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
				typ = typ.Elem()
			}
		case reflect.Array, reflect.Slice:
			typ = typ.Elem()
		}

		switch typ.Kind() {
		// Basically all types which cannot be unmarshalled by the json package.
		case reflect.Chan, reflect.Complex64, reflect.Complex128, reflect.Func, reflect.Interface:
			return fmt.Errorf("jsonapi: attribute %q of type %q is of unsupported type",
				sf.Name,
				resType,
			)
		}
	}

	// Check relationships
	for i := 0; i < value.NumField(); i++ {
		sf := value.Type().Field(i)

		if strings.HasPrefix(sf.Tag.Get("api"), "rel,") {
			s := strings.Split(sf.Tag.Get("api"), ",")

			if len(s) < 2 || len(s) > 3 {
				return fmt.Errorf(
					"jsonapi: api tag of relationship %q of struct %q is invalid",
					sf.Name,
					value.Type().Name(),
				)
			}

			if sf.Type.String() != "string" && sf.Type.String() != "[]string" {
				return fmt.Errorf(
					"jsonapi: relationship %q of type %q is not string or []string",
					sf.Name,
					resType,
				)
			}
		}
	}

	return nil
}

// BuildType takes a struct or a pointer to a struct to analyse and builds a
// Type object that is returned.
//
// If an error is returned, the Type object will be empty.
func BuildType(v interface{}) (Type, error) {
	typ := Type{}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return typ, errors.New("jsonapi: value must represent a struct")
	}

	err := Check(val.Interface())
	if err != nil {
		return typ, fmt.Errorf("jsonapi: invalid type: %q", err)
	}

	typ.Name, typ.Attrs, typ.Rels = getTypeInfo(val)

	// NewFunc
	res := Wrap(reflect.New(val.Type()).Interface())
	typ.NewFunc = res.Copy

	return typ, nil
}

// MustBuildType calls BuildType and returns the result.
//
// It panics if the error returned by BuildType is not nil.
func MustBuildType(v interface{}) Type {
	typ, err := BuildType(v)
	if err != nil {
		panic(err)
	}

	return typ
}

func getTypeInfo(val reflect.Value) (string, map[string]Attr, map[string]Rel) {
	idSF, _ := val.Type().FieldByName("ID")
	typeName := idSF.Tag.Get("api")

	attrs := map[string]Attr{}

	for i := 0; i < val.NumField(); i++ {
		fs := val.Type().Field(i)
		jsonTag := fs.Tag.Get("json")
		apiTag := fs.Tag.Get("api")

		attr := strings.Split(apiTag, ",")
		if attr[0] == "attr" {
			typ, arr, null := GetAttrType(fs.Type.String())

			if len(attr) >= 2 {
				// If the attribute type is not registered, typ equals 0, which is the same
				// as AttrTypeInvalid.
				typ = registry.namesR[attr[1]]
			}

			if len(attr) >= 3 {
				arr = arr && attr[2] != "no-array"
			}

			attrs[jsonTag] = Attr{
				Name:     jsonTag,
				Type:     typ,
				Array:    arr,
				Nullable: null,
			}
		}
	}

	// Relationships
	rels := map[string]Rel{}

	for i := 0; i < val.NumField(); i++ {
		fs := val.Type().Field(i)
		jsonTag := fs.Tag.Get("json")
		relTag := strings.Split(fs.Tag.Get("api"), ",")
		invName := ""

		if len(relTag) == 3 {
			invName = relTag[2]
		}

		toOne := true
		if fs.Type.String() == "[]string" {
			toOne = false
		}

		if relTag[0] == "rel" {
			rels[jsonTag] = Rel{
				FromName: jsonTag,
				ToType:   relTag[1],
				ToOne:    toOne,
				ToName:   invName,
				FromType: typeName,
			}
		}
	}

	return typeName, attrs, rels
}

// ReduceRels removes redundant relationship paths.
//
// DO NOT use this for non-static relationship paths, such as filters or inclusions.
func ReduceRels(rels []Rel) []Rel {
	r := make([]Rel, len(rels))
	copy(r, rels)

	for i := len(r) - 1; i >= 0; i-- {
		for j := 0; j <= i; j++ {
			if r[j].FromType == r[i].ToType {
				r = append(r[:j], r[i+1:]...)
				i = j

				break
			}
		}
	}

	return r
}
