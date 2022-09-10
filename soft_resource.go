package jsonapi

import (
	"fmt"
	"strings"
	"time"
)

// SoftResource represents a resource whose type is defined by an internal field
// of type *Type.
//
// Changing the type automatically changes the resource's attributes and
// relationships. When a field is added, its value is the zero value of the
// field's type.
type SoftResource struct {
	Type *Type

	id   string
	data map[string]interface{}
	meta Meta
}

// Attrs returns the resource's attributes.
func (sr *SoftResource) Attrs() map[string]Attr {
	sr.check()

	return sr.Type.Attrs
}

// Rels returns the resource's relationships.
func (sr *SoftResource) Rels() map[string]Rel {
	sr.check()

	return sr.Type.Rels
}

// AddAttr adds an attribute.
func (sr *SoftResource) AddAttr(attr Attr) {
	sr.check()

	for _, name := range sr.fields() {
		if name == attr.Name {
			return
		}
	}

	sr.Type.Attrs[attr.Name] = attr
}

// AddRel adds a relationship.
func (sr *SoftResource) AddRel(rel Rel) {
	sr.check()

	for _, name := range sr.fields() {
		if name == rel.FromName {
			return
		}
	}

	sr.Type.Rels[rel.FromName] = rel
}

// RemoveField removes a field.
func (sr *SoftResource) RemoveField(field string) {
	sr.check()
	delete(sr.Type.Attrs, field)
	delete(sr.Type.Rels, field)
}

// Attr returns the attribute named after key.
func (sr *SoftResource) Attr(key string) Attr {
	sr.check()

	return sr.Type.Attrs[key]
}

// Rel returns the relationship named after key.
func (sr *SoftResource) Rel(key string) Rel {
	sr.check()

	return sr.Type.Rels[key]
}

// New returns a new resource (of type SoftResource) with the same type but
// without the values.
func (sr *SoftResource) New() Resource {
	sr.check()

	typ := sr.Type.Copy()

	return &SoftResource{
		Type: &typ,
	}
}

// GetID returns the resource's ID.
func (sr *SoftResource) GetID() string {
	sr.check()

	return sr.id
}

// GetType returns the resource's type.
func (sr *SoftResource) GetType() Type {
	sr.check()

	return *sr.Type
}

// Get returns the value associated to the field named after key.
func (sr *SoftResource) Get(key string) interface{} {
	sr.check()

	if key == "id" {
		return sr.GetID()
	}

	if _, ok := sr.Type.Attrs[key]; ok {
		if v, ok := sr.data[key]; ok {
			return v
		}
	} else if _, ok := sr.Type.Rels[key]; ok {
		if v, ok := sr.data[key]; ok {
			return v
		}
	}

	return nil
}

// SetID sets the resource's ID.
func (sr *SoftResource) SetID(id string) {
	sr.check()
	sr.id = id
}

// SetType sets the resource's type.
func (sr *SoftResource) SetType(typ *Type) {
	sr.check()
	sr.Type = typ
}

// Set sets the value associated to the field named key to v.
func (sr *SoftResource) Set(key string, v interface{}) {
	sr.check()

	if key == "id" {
		id, _ := v.(string)
		sr.id = id

		return
	}

	if attr, ok := sr.Type.Attrs[key]; ok {
		if isNil(v) {
			if attr.Unmarshaler != nil {
				sr.data[key] = attr.Unmarshaler.GetZeroValue(attr.Array, attr.Nullable)
			} else {
				sr.data[key] = GetZeroValue(attr.Type, attr.Array, attr.Nullable)
			}

			return
		}

		// byte is an alias for uint8, so %T will return uint8.
		t := fmt.Sprintf("%T", v)
		if attr.Type == AttrTypeBytes && strings.HasSuffix(t, "uint8") {
			t = strings.Replace(t, "uint8", "byte", 1)
		}

		if attr.Unmarshaler != nil {
			ok, arr, null := attr.Unmarshaler.CheckAttrType(t)
			if ok && attr.Array == arr && attr.Nullable == null {
				sr.data[key] = v
			}
		} else {
			typ, arr, null := GetAttrType(t)
			if (attr.Type == typ && attr.Array == arr && attr.Nullable == null) ||
				(attr.Type == AttrTypeBytes && arr && attr.Nullable == null) {
				sr.data[key] = v
			}
		}
	} else if rel, ok := sr.Type.Rels[key]; ok {
		if _, ok := v.(string); ok && rel.ToOne {
			sr.data[key] = v
		} else if _, ok := v.([]string); ok && !rel.ToOne {
			sr.data[key] = v
		}
	}
}

// Copy returns a new SoftResource object with the same type and values.
func (sr *SoftResource) Copy() Resource {
	sr.check()

	typ := sr.Type.Copy()

	return &SoftResource{
		Type: &typ,
		id:   sr.id,
		data: copyData(sr.data),
	}
}

// Meta returns the meta values of the resource.
func (sr *SoftResource) Meta() Meta {
	return sr.meta
}

// SetMeta sets the meta values of the resource.
func (sr *SoftResource) SetMeta(m Meta) {
	sr.meta = m
}

func (sr *SoftResource) fields() []string {
	fields := make([]string, 0, len(sr.Type.Attrs)+len(sr.Type.Rels))
	for i := range sr.Type.Attrs {
		fields = append(fields, sr.Type.Attrs[i].Name)
	}

	for i := range sr.Type.Rels {
		fields = append(fields, sr.Type.Rels[i].FromName)
	}

	return fields
}

func (sr *SoftResource) check() {
	if sr.Type == nil {
		sr.Type = &Type{}
	}

	if sr.Type.Attrs == nil {
		sr.Type.Attrs = map[string]Attr{}
	}

	if sr.Type.Rels == nil {
		sr.Type.Rels = map[string]Rel{}
	}

	if sr.data == nil {
		sr.data = map[string]interface{}{}
	}

	for i := range sr.Type.Attrs {
		attr := sr.Type.Attrs[i]

		if _, ok := sr.data[attr.Name]; !ok {
			if attr.Type == AttrTypeBytes {
				sr.data[attr.Name] = GetZeroValue(attr.Type, true, attr.Nullable)
			} else {
				sr.data[attr.Name] = GetZeroValue(attr.Type, attr.Array, attr.Nullable)
			}
		}
	}

	for i := range sr.Type.Rels {
		n := sr.Type.Rels[i].FromName
		if _, ok := sr.data[n]; !ok {
			if sr.Type.Rels[i].ToOne {
				sr.data[n] = ""
			} else {
				sr.data[n] = []string{}
			}
		}
	}

	fields := sr.fields()
	if len(fields) < len(sr.data) {
		for k := range sr.data {
			found := false

			for _, f := range fields {
				if k == f {
					found = true
					break
				}
			}

			if !found {
				delete(sr.data, k)
			}
		}
	}
}

func copyData(d map[string]interface{}) map[string]interface{} {
	// todo: handle AttrTypeOther
	d2 := map[string]interface{}{}

	for k, v := range d {
		switch v2 := v.(type) {
		// String array
		case []string:
			nv := make([]string, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]string:
			if v2 == nil {
				d2[k] = (*[]string)(nil)
			} else {
				nv := make([]string, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Int array
		case []int:
			nv := make([]int, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]int:
			if v2 == nil {
				d2[k] = (*[]int)(nil)
			} else {
				nv := make([]int, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Int8 array
		case []int8:
			nv := make([]int8, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]int8:
			if v2 == nil {
				d2[k] = (*[]int8)(nil)
			} else {
				nv := make([]int8, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Int16 array
		case []int16:
			nv := make([]int16, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]int16:
			if v2 == nil {
				d2[k] = (*[]int16)(nil)
			} else {
				nv := make([]int16, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Int32 array
		case []int32:
			nv := make([]int32, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]int32:
			if v2 == nil {
				d2[k] = (*[]int32)(nil)
			} else {
				nv := make([]int32, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Int64 array
		case []int64:
			nv := make([]int64, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]int64:
			if v2 == nil {
				d2[k] = (*[]int64)(nil)
			} else {
				nv := make([]int64, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Uint array
		case []uint:
			nv := make([]uint, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]uint:
			if v2 == nil {
				d2[k] = (*[]uint)(nil)
			} else {
				nv := make([]uint, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Uint8 array
		case []uint8:
			nv := make([]uint8, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]uint8:
			if v2 == nil {
				d2[k] = (*[]uint8)(nil)
			} else {
				nv := make([]uint8, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Uint16 array
		case []uint16:
			nv := make([]uint16, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]uint16:
			if v2 == nil {
				d2[k] = (*[]uint16)(nil)
			} else {
				nv := make([]uint16, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Uint32 array
		case []uint32:
			nv := make([]uint32, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]uint32:
			if v2 == nil {
				d2[k] = (*[]uint32)(nil)
			} else {
				nv := make([]uint32, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Uint64 array
		case []uint64:
			nv := make([]uint64, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]uint64:
			if v2 == nil {
				d2[k] = (*[]uint64)(nil)
			} else {
				nv := make([]uint64, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Bool array
		case []bool:
			nv := make([]bool, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]bool:
			if v2 == nil {
				d2[k] = (*[]bool)(nil)
			} else {
				nv := make([]bool, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		// Time array
		case []time.Time:
			nv := make([]time.Time, len(v2))
			_ = copy(nv, v2)
			d2[k] = v2
		case *[]time.Time:
			if v2 == nil {
				d2[k] = (*[]time.Time)(nil)
			} else {
				nv := make([]time.Time, len(*v2))
				_ = copy(nv, *v2)
				d2[k] = v2
			}
		default:
			d2[k] = v2
		}
	}

	return d2
}
