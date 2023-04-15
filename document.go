package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

// A Document represents a JSON:API document.
type Document struct {
	// Data
	Data interface{}

	// Included
	Included []Resource

	// References
	Resources map[string]map[string]struct{}
	Links     map[string]Link

	// Relationships where data has to be included in payload
	RelData map[string][]string

	// Top-level members
	Meta Meta

	// Errors
	Errors []Error

	// Internal
	PrePath string
}

// Include adds res to the set of resources to be included under the included
// top-level field.
//
// It also makes sure that resources are not added twice.
func (d *Document) Include(res Resource) {
	key := res.Get("id").(string) + " " + res.GetType().Name

	if len(d.Included) == 0 {
		d.Included = []Resource{}
	}

	if dres, ok := d.Data.(Resource); ok {
		// Check resource
		rkey := dres.Get("id").(string) + " " + dres.GetType().Name

		if rkey == key {
			return
		}
	} else if col, ok := d.Data.(Collection); ok {
		// Check Collection
		ctyp := col.GetType()
		if ctyp.Name == res.GetType().Name {
			for i := 0; i < col.Len(); i++ {
				rkey := col.At(i).Get("id").(string) + " " + col.At(i).GetType().Name

				if rkey == key {
					return
				}
			}
		}
	}

	// Check already included resources
	for _, res := range d.Included {
		if key == res.Get("id").(string)+" "+res.GetType().Name {
			return
		}
	}

	d.Included = append(d.Included, res)
}

// MarshalDocument marshals a document according to the JSON:API specification.
//
// Both doc and url must not be nil.
func MarshalDocument(dst io.Writer, doc *Document, url *URL) error {
	var err error

	// Data
	var data json.RawMessage

	switch d := doc.Data.(type) {
	case Resource:
		if url.Params.Fields != nil {
			data = MarshalResource(d, doc.PrePath, url.Params.Fields[d.GetType().Name], doc.RelData)
		} else {
			data = MarshalResource(d, doc.PrePath, nil, doc.RelData)
		}
	case Collection:
		data = MarshalCollection(
			d,
			doc.PrePath,
			url.Params.Fields,
			doc.RelData,
		)
	case Identifier:
		data, err = json.Marshal(d)
	case Identifiers:
		data, err = json.Marshal(d)
	default:
		if doc.Data != nil {
			err = errors.New("data contains an unknown type")
		} else if len(doc.Errors) == 0 {
			data = []byte("null")
		}
	}

	// Data
	var errs json.RawMessage
	if len(doc.Errors) > 0 {
		// Errors
		errs, err = json.Marshal(doc.Errors)
	}

	if err != nil {
		return err
	}

	// Included
	var inclusions []*json.RawMessage

	if len(doc.Included) > 0 {
		sort.Slice(doc.Included, func(i, j int) bool {
			return doc.Included[i].Get("id").(string) < doc.Included[j].Get("id").(string)
		})

		if len(data) > 0 {
			for key := range doc.Included {
				typ := doc.Included[key].GetType().Name
				raw := MarshalResource(
					doc.Included[key],
					doc.PrePath,
					url.Params.Fields[typ],
					doc.RelData,
				)
				rawm := json.RawMessage(raw)
				inclusions = append(inclusions, &rawm)
			}
		}
	}

	// Marshaling
	plMap := map[string]interface{}{}

	if len(errs) > 0 {
		plMap["errors"] = errs
	} else if len(data) > 0 {
		plMap["data"] = data

		if len(inclusions) > 0 {
			plMap["included"] = inclusions
		}
	}

	if len(doc.Meta) > 0 {
		plMap["meta"] = doc.Meta
	}

	links := doc.Links

	if url != nil {
		if links == nil {
			links = map[string]Link{}
		}

		links["self"] = Link{
			HRef: doc.PrePath + url.String(),
		}
	}

	if links != nil {
		plMap["links"] = links
	}

	plMap["jsonapi"] = map[string]string{"version": "1.0"}

	return json.NewEncoder(dst).Encode(plMap)
}

var (
	errMissingPrimaryMember = errors.New("jsonapi: missing primary member")
	errCoexistingMembers    = errors.New(`jsonapi: "data" and "errors" must not coexist`)
	errMemberDataType       = errors.New("jsonapi: invalid member data type")
)

// UnmarshalDocument reads a payload to build and return a Document object.
//
// schema must not be nil.
func UnmarshalDocument(r io.Reader, schema *Schema) (*Document, error) {
	doc := &Document{
		Included:  []Resource{},
		Resources: map[string]map[string]struct{}{},
		Links:     map[string]Link{},
		RelData:   map[string][]string{},
		Meta:      map[string]interface{}{},
	}
	ske := &payloadSkeleton{}
	dec := json.NewDecoder(r)

	// Unmarshal
	if err := dec.Decode(ske); err != nil {
		return nil, payloadErr(err)
	}

	// SPEC 5.1
	// A document MUST contain at least one of the following three members.
	if ske.Data == nil && ske.Errors == nil && ske.Meta == nil {
		return nil, payloadErr(errMissingPrimaryMember)
	}

	// SPEC 5.1
	// data and errors must not coexist.
	if ske.Data != nil && ske.Errors != nil {
		return nil, payloadErr(errCoexistingMembers)
	}

	// Data
	switch {
	case len(ske.Data) > 0:
		switch {
		case ske.Data[0] == '{':
			// Resource
			res, err := UnmarshalResource(ske.Data, schema)
			if err != nil {
				return nil, &srcError{
					ptr:   true,
					src:   "/data",
					error: fmt.Errorf("jsonapi: failed to unmarshal resource: %w", err),
				}
			}

			doc.Data = res
		case ske.Data[0] == '[':
			col, err := UnmarshalCollection(ske.Data, schema)
			if err != nil {
				return nil, &srcError{
					ptr:   true,
					src:   "/data",
					error: fmt.Errorf("jsonapi: failed to unmarshal collection: %w", err),
				}
			}

			doc.Data = col
		case string(ske.Data) == "null":
			doc.Data = nil
		default:
			return nil, &srcError{ptr: true, src: "/data", error: payloadErr(errMemberDataType)}
		}
	case len(ske.Errors) > 0:
		doc.Errors = ske.Errors
	}

	// Included
	for i, raw := range ske.Included {
		res, err := UnmarshalResource(raw, schema)
		if err != nil {
			return nil, fmt.Errorf("jsonapi: failed to unmarshal included resource at %d: %w",
				i, &srcError{src: fmt.Sprintf("/included/%d", i), ptr: true, error: err})
		}

		doc.Included = append(doc.Included, res)
	}

	// Meta
	doc.Meta = ske.Meta

	return doc, nil
}
