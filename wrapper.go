package jsonapi

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Wrapper wraps a reflect.Value that represents a struct.
//
// The Wrap function can be used to wrap a struct and make a Wrapper object.
//
// It implements the Resource interface, so the value can be handled as if it
// were a Resource.
type Wrapper struct {
	val  reflect.Value // Actual value (with content)
	typ  Type
	meta Meta
}

// Wrap wraps v (a struct or a pointer to a struct) and returns a Wrapper that
// can be used as a Resource to handle the given value.
//
// Changes made to the Wrapper object (through Set for example) will be applied
// to v.
//
// If v is not a pointer, a copy is made and v won't be modified by the wrapper.
func Wrap(v interface{}) *Wrapper {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr {
		newVal := reflect.New(val.Type()).Elem()
		for i := 0; i < newVal.NumField(); i++ {
			f := newVal.Field(i)
			if f.CanSet() {
				f.Set(val.Field(i))
			}
		}

		val = newVal
	} else {
		val = val.Elem()
	}

	err := Check(val.Interface())
	if err != nil {
		panic("invalid struct: " + err.Error())
	}

	typ, attrs, rels := getTypeInfo(val)
	for _, attr := range attrs {
		if attr.Type == AttrTypeInvalid {
			panic(fmt.Sprintf("jsonapi: unable to resolve attribute type for \"%s.%s\"",
				typ, attr.Name))
		}
	}

	w := &Wrapper{
		val: val,
		typ: Type{
			Name:  typ,
			Attrs: attrs,
			Rels:  rels,
		},
	}

	// Meta
	if m, ok := v.(MetaHolder); ok {
		if len(m.Meta()) > 0 {
			w.SetMeta(m.Meta())
		}
	}

	return w
}

// IDAndType returns the ID and the type of the Wrapper.
func (w *Wrapper) IDAndType() (string, string) {
	return IDAndType(w.val.Interface())
}

// Attrs returns the attributes of the Wrapper.
func (w *Wrapper) Attrs() map[string]Attr {
	return w.typ.Attrs
}

// Rels returns the relationships of the Wrapper.
func (w *Wrapper) Rels() map[string]Rel {
	return w.typ.Rels
}

// Attr returns the attribute that corresponds to the given key.
func (w *Wrapper) Attr(key string) Attr {
	for _, attr := range w.typ.Attrs {
		if attr.Name == key {
			return attr
		}
	}

	return Attr{}
}

// Rel returns the relationship that corresponds to the given key.
func (w *Wrapper) Rel(key string) Rel {
	for _, rel := range w.typ.Rels {
		if rel.FromName == key {
			return rel
		}
	}

	return Rel{}
}

// New returns a copy of the resource under the wrapper.
func (w *Wrapper) New() Resource {
	newVal := reflect.New(w.val.Type())
	return Wrap(newVal.Interface())
}

// GetID returns the wrapped resource's ID.
func (w *Wrapper) GetID() string {
	id, _ := IDAndType(w.val.Interface())
	return id
}

// GetType returns the wrapped resource's type.
func (w *Wrapper) GetType() Type {
	return w.typ
}

// Get returns the value associated to the attribute named after key.
func (w *Wrapper) Get(key string) interface{} {
	if key == "id" {
		return w.GetID()
	}

	return w.getField(key)
}

// SetID sets the ID of the wrapped resource.
func (w *Wrapper) SetID(id string) {
	w.val.FieldByName("ID").SetString(id)
}

// Set sets the value associated to the attribute named after key.
func (w *Wrapper) Set(key string, val interface{}) {
	if key == "id" {
		id, _ := val.(string)
		w.SetID(id)
	}

	w.setField(key, val)
}

// Copy makes a copy of the wrapped resource and returns it.
//
// The returned value's concrete type is also a Wrapper.
func (w *Wrapper) Copy() Resource {
	nw := Wrap(reflect.New(w.val.Type()).Interface())

	// Attributes
	for _, attr := range w.Attrs() {
		nw.Set(attr.Name, w.Get(attr.Name))
	}

	// Relationships
	for _, rel := range w.Rels() {
		if rel.ToOne {
			nw.Set(rel.FromName, w.Get(rel.FromName).(string))
		} else {
			nw.Set(rel.FromName, w.Get(rel.FromName).([]string))
		}
	}

	return nw
}

// Meta returns the meta values of the resource.
func (w *Wrapper) Meta() Meta {
	return w.meta
}

// SetMeta sets the meta values of the resource.
func (w *Wrapper) SetMeta(m Meta) {
	w.meta = m
}

// Private methods

func (w *Wrapper) getField(key string) interface{} {
	if key == "" {
		panic("key is empty")
	}

	for i := 0; i < w.val.NumField(); i++ {
		field := w.val.Field(i)
		sf := w.val.Type().Field(i)

		if key == sf.Tag.Get("json") && sf.Tag.Get("api") != "" {
			// If a key does not exist in the attribute map, it's an relationship and does not have
			// a "zero value".
			if attr, ok := w.typ.Attrs[key]; ok && isNil(field.Interface()) {
				zv, _ := GetZeroValue(attr.Type, attr.Array, attr.Nullable)
				return zv
			}

			return field.Interface()
		}
	}

	panic(fmt.Sprintf("attribute %q does not exist", key))
}

func (w *Wrapper) setField(key string, v interface{}) {
	if key == "" {
		panic("key is empty")
	}

	for i := 0; i < w.val.NumField(); i++ {
		field := w.val.Field(i)
		sf := w.val.Type().Field(i)

		if key == sf.Tag.Get("json") {
			if v == nil {
				field.Set(reflect.New(field.Type()).Elem())
				return
			}

			val := reflect.ValueOf(v)
			if val.Type() == field.Type() {
				field.Set(val)
				return
			}

			panic(fmt.Sprintf(
				"got value of type %q, not %q",
				field.Type(), val.Type(),
			))
		}
	}

	panic(fmt.Sprintf("attribute %q does not exist", key))
}

// ReflectTypeUnmarshaler is a reflection based type unmarshaler.
type ReflectTypeUnmarshaler struct {
	// Type is the "base" type (not nullable and not an array) of the attribute. For example, for
	// a Struct property of type "*[]string", reflect.TypeOf("") would be correct.
	Type reflect.Type
}

func (u ReflectTypeUnmarshaler) typ(array, nullable bool) reflect.Type {
	switch {
	case array && nullable:
		return reflect.New(reflect.SliceOf(u.Type)).Type()
	case array:
		return reflect.SliceOf(u.Type)
	case nullable:
		return reflect.New(u.Type).Type()
	}

	return u.Type
}

func (u ReflectTypeUnmarshaler) GetZeroValue(_ int, array, nullable bool) interface{} {
	var typ reflect.Type

	switch {
	case array && nullable:
		typ = u.typ(true, true)
		return reflect.Zero(typ).Interface()
	case array:
		typ = u.typ(true, false)
		return reflect.MakeSlice(typ, 0, 0).Interface()
	case nullable:
		typ = u.typ(false, true)
		return reflect.New(typ).Elem().Interface()
	}

	typ = u.Type

	// Make sure that types with nil-zero value are initialized empty instead.
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		return reflect.MakeSlice(typ, 0, 0).Slice(0, 0).Interface()
	} else if typ.Kind() == reflect.Map {
		return reflect.MakeMap(typ).Interface()
	}

	return reflect.Zero(typ).Interface()
}

func (u ReflectTypeUnmarshaler) UnmarshalToType(data []byte, attr Attr) (interface{}, error) {
	if data == nil || (!attr.Nullable && string(data) == "null") {
		return nil, fmt.Errorf("type is not nullable")
	}

	if attr.Nullable && string(data) == "null" {
		return u.GetZeroValue(attr.Type, attr.Array, attr.Nullable), nil
	}

	var (
		val interface{}
		err error
	)

	if attr.Array {
		tv := reflect.New(u.typ(true, false))
		err = json.Unmarshal(data, tv.Interface())

		if attr.Nullable {
			val = tv.Interface()
		} else {
			val = tv.Elem().Interface()
		}
	} else {
		tv := reflect.New(u.typ(false, false))
		err = json.Unmarshal(data, tv.Interface())

		if attr.Nullable {
			val = tv.Interface()
		} else {
			val = tv.Elem().Interface()
		}
	}

	if err != nil {
		return nil, err
	}

	return val, nil
}
