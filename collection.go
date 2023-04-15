package jsonapi

import (
	"encoding/json"
	"fmt"
)

// A Collection defines the interface of a structure that can manage a set of
// ordered resources of the same type.
type Collection interface {
	// GetType returns the type of the resources.
	GetType() Type

	// Len returns the number of resources in the collection.
	Len() int

	// At returns the resource at index i.
	At(int) Resource

	// Add adds a resource in the collection.
	Add(Resource)
}

// MarshalCollection marshals a Collection into a JSON-encoded payload.
func MarshalCollection(c Collection, prepath string, fields map[string][]string, relData map[string][]string) []byte {
	var raws []*json.RawMessage

	if c.Len() == 0 {
		return []byte("[]")
	}

	for i := 0; i < c.Len(); i++ {
		r := c.At(i)
		raw := json.RawMessage(
			MarshalResource(r, prepath, fields[r.GetType().Name], relData),
		)
		raws = append(raws, &raw)
	}

	// NOTE An error should not happen.
	pl, _ := json.Marshal(raws)

	return pl
}

// UnmarshalCollection unmarshals a JSON-encoded payload into a Collection.
func UnmarshalCollection(data []byte, schema *Schema) (Collection, error) {
	var cske []json.RawMessage

	err := json.Unmarshal(data, &cske)
	if err != nil {
		return nil, payloadErr(err)
	}

	col := &Resources{}

	for i := range cske {
		res, err := UnmarshalResource(cske[i], schema)
		if err != nil {
			return nil, fmt.Errorf("jsonapi: failed to unmarshal resource at %d: %w",
				i, &srcError{src: fmt.Sprintf("/%d", i), ptr: true, error: err})
		}

		col.Add(res)
	}

	return col, nil
}

// Resources is a slice of objects that implements the Collection interface. The resources
// do not necessarily have to be of the same type.
type Resources []Resource

// GetType returns a zero Type object because the collection does not represent
// a particular type.
func (r *Resources) GetType() Type {
	return Type{}
}

// Len returns the number of elements in the collection.
func (r *Resources) Len() int {
	return len(*r)
}

// At returns the resource at position i. If the index is out of bounds, nil is returned.
func (r *Resources) At(i int) Resource {
	if i >= 0 && i < r.Len() {
		return (*r)[i]
	}

	return nil
}

// Add adds a Resource object to the collection.
func (r *Resources) Add(res Resource) {
	*r = append(*r, res)
}
