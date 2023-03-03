package jsonapi

import (
	"sort"
	"strings"
)

// NewParams creates and returns a Params object built from a SimpleURL and a
// given resource type. A schema is used for validation.
//
// If validation is not expected, it is recommended to simply build a SimpleURL
// object with NewSimpleURL.
func NewParams(schema *Schema, su SimpleURL, resType string) (*Params, error) {
	params := &Params{}

	// Include
	if su.Include != nil {
		incs := make([]string, len(su.Include))
		copy(incs, su.Include)
		sort.Strings(incs)

		// Remove duplicates and unnecessary includes
		for i := len(incs) - 1; i >= 0; i-- {
			if i > 0 && strings.HasPrefix(incs[i], incs[i-1]) {
				incs = append(incs[:i-1], incs[i:]...)
			}
		}

		// Check inclusions
		params.Include = make([][]Rel, len(incs))

		for i := 0; i < len(incs); i++ {
			words := strings.Split(incs[i], ".")
			params.Include[i] = make([]Rel, len(words))

			// incRel and typ are overridden for each "part" of the relationship path.
			incRel := Rel{ToType: resType}
			for j, word := range words {
				typ := schema.GetType(incRel.ToType)

				var ok bool
				if incRel, ok = typ.Rels[word]; !ok {
					return nil, NewErrInvalidRelationshipPath(incRel.ToType, word)
				}

				// For each resource encountered in the multipart path, an empty fields slice added
				// so the fields can be populated with default fields if needed.
				// SPEC (v1.0) 6.3, second note block
				params.Include[i][j] = incRel
			}
		}
	}

	// Fields
	if len(su.Fields) != 0 {
		// After these checks, only valid fields remain, representing either the resource ID or
		// one of the attributes or relations.
		for typeName, fields := range su.Fields {
			fields = removeDuplicates(fields)

			typ := schema.GetType(typeName)
			if typeName != resType && typ.Name == "" {
				return nil, NewErrUnknownTypeInURL(typeName)
			}

			// Check if the sparse fieldset contains any fields that does not exist on the type.
			if field := findFirstDifference(fields, typ.Fields()); field != "" && field != "id" {
				return nil, NewErrUnknownFieldInURL(field)
			}

			if len(params.Fields) == 0 {
				params.Fields = map[string][]string{}
			}

			params.Fields[typeName] = fields
		}

		// Separate the passed fields into attributes and relationships.
		for typeName, fields := range params.Fields {
			// This should always return a type since
			// it is checked earlier.
			typ := schema.GetType(typeName)

			rels := make([]Rel, 0, len(typ.Attrs))
			attrs := make([]Attr, 0, len(typ.Attrs))

			// Get Attrs and Rels for the requested fields
			for _, field := range typ.Fields() {
				for _, field2 := range fields {
					if field != field2 {
						continue
					}

					typ = schema.GetType(typeName)
					if typ.Name == "" {
						continue
					}

					if attr, ok := typ.Attrs[field]; ok {
						// Append to list of attributes
						attrs = append(attrs, attr)
					} else if rel, ok := typ.Rels[field]; ok {
						// Append to list of relationships
						rels = append(rels, rel)
					}
				}
			}

			if len(attrs) != 0 {
				if len(params.Attrs) == 0 {
					params.Attrs = map[string][]Attr{}
				}

				params.Attrs[typeName] = attrs
			}

			if len(rels) != 0 {
				if len(params.Rels) == 0 {
					params.Rels = map[string][]Rel{}
				}

				params.Rels[typeName] = rels
			}
		}
	}

	// Sorting
	// TODO All of the following is just to figure out
	// if the URL represents a single resource or a
	// collection. It should be done in a better way.
	isCol := false
	if len(su.Fragments) == 1 {
		isCol = true
	} else if len(su.Fragments) >= 3 {
		relName := su.Fragments[len(su.Fragments)-1]
		typ := schema.GetType(su.Fragments[0])
		// Checked earlier, assuming should be safe
		rel := typ.Rels[relName]
		isCol = !rel.ToOne
	}

	if isCol {
		typ := schema.GetType(resType)

		// Check for conflicting sort fields.
		set := make(map[string]struct{}, len(su.SortingRules))

		for _, sr := range su.SortingRules {
			sr = strings.TrimPrefix(sr, "-")
			if _, ok := set[sr]; ok {
				return nil, NewErrConflictingSortField(sr)
			}

			set[sr] = struct{}{}
		}

		for _, rule := range su.SortingRules {
			sr := SortRule{}

			if rule[0] == '-' {
				rule = rule[1:]
				sr.Desc = true
			}

			parts := strings.Split(rule, ".")
			if len(parts) == 1 {
				sr.Name = parts[0]
				if _, ok := typ.Attrs[sr.Name]; !ok && sr.Name != "id" {
					return nil, NewErrUnknownSortField(typ.Name, sr.Name)
				}

				params.SortRules = append(params.SortRules, sr)

				continue
			}

			var path []Rel

			st := typ
			for i := 0; i < len(parts)-1; i++ {
				rel, ok := st.Rels[parts[i]]
				if !ok || !rel.ToOne {
					return nil, NewErrUnknownSortRelationship(st.Name, parts[i])
				}

				path = append(path, rel)
				st = schema.GetType(rel.ToType)
			}

			sr.Name = parts[len(parts)-1]
			if _, ok := st.Attrs[sr.Name]; !ok && sr.Name != "id" {
				return nil, NewErrUnknownSortField(st.Name, sr.Name)
			}

			// By reducing the relationship path, we may be able to eliminate unnecessary
			// nodes.
			path = ReduceRels(path)
			if len(path) != 0 {
				sr.Path = path
			} else {
				sr.Path = nil
			}

			params.SortRules = append(params.SortRules, sr)
		}
	}

	// Filter
	params.Filter = su.Filter

	// Pagination
	params.Page = su.Page

	// Off-Spec query params
	params.Params = su.Params

	return params, nil
}

// removeDuplicates creates a sorted copy of s without duplicates.
func removeDuplicates(s []string) []string {
	s2 := make([]string, len(s))
	copy(s2, s)
	sort.Strings(s2)

	for i := len(s2) - 1; i >= 0; i-- {
		if i > 0 && s2[i] == s2[i-1] {
			s2 = append(s2[:i-1], s2[i:]...)
		}
	}

	return s2
}

func findFirstDifference(a, b []string) string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}

	for _, x := range a {
		if _, found := mb[x]; !found {
			return x
		}
	}

	return ""
}

// Params represents all the query parameters from the URL.
type Params struct {
	// Fields contains the names of all attributes and relationships that are included in the
	// sparse field sets.
	Fields map[string][]string

	// Attrs contains all attributes found in Fields, grouped by the name of the resource type
	// they belong to.
	Attrs map[string][]Attr

	// Rels contains all relationships found in Fields, grouped by the name of the resource type
	// they belong to.
	Rels map[string][]Rel

	// Filter
	Filter map[string][]string

	// SortRules contains all sorting rules.
	SortRules []SortRule

	// Pagination
	Page map[string]string

	// Include contains cleaned up relationship paths.
	Include [][]Rel

	// Params contains all off-spec query parameters.
	Params map[string][]string
}
