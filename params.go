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
	params := &Params{
		Fields:    map[string][]string{},
		Attrs:     map[string][]Attr{},
		Rels:      map[string][]Rel{},
		Filter:    nil,
		SortRules: []SortRule{},
		Include:   [][]Rel{},
	}

	// Include
	incs := make([]string, len(su.Include))
	copy(incs, su.Include)
	sort.Strings(incs)

	// Remove duplicates and unnecessary includes
	for i := len(incs) - 1; i >= 0; i-- {
		if i > 0 {
			if strings.HasPrefix(incs[i], incs[i-1]) {
				incs = append(incs[:i-1], incs[i:]...)
			}
		}
	}

	// Check inclusions
	for i := 0; i < len(incs); i++ {
		words := strings.Split(incs[i], ".")

		incRel := Rel{ToType: resType}

		for _, word := range words {
			if typ := schema.GetType(incRel.ToType); typ.Name != "" {
				var ok bool
				if incRel, ok = typ.Rels[word]; ok {
					params.Fields[incRel.ToType] = []string{}
				} else {
					incs = append(incs[:i], incs[i+1:]...)
					break
				}
			}
		}
	}

	// Build params.Include
	params.Include = make([][]Rel, len(incs))

	for i := range incs {
		words := strings.Split(incs[i], ".")

		params.Include[i] = make([]Rel, len(words))

		var incRel Rel

		for w := range words {
			if w == 0 {
				typ := schema.GetType(resType)
				incRel = typ.Rels[words[0]]
			}

			params.Include[i][w] = incRel

			if w < len(words)-1 {
				typ := schema.GetType(incRel.ToType)
				incRel = typ.Rels[words[w+1]]
			}
		}
	}

	if resType != "" {
		params.Fields[resType] = []string{}
	}

	// Fields
	for t, fields := range su.Fields {
		if t != resType {
			if typ := schema.GetType(t); typ.Name == "" {
				return nil, NewErrUnknownTypeInURL(t)
			}
		}

		if typ := schema.GetType(t); typ.Name != "" {
			params.Fields[t] = []string{}

			for _, f := range fields {
				if f == "id" {
					params.Fields[t] = append(params.Fields[t], "id")
				} else {
					for _, ff := range typ.Fields() {
						if f == ff {
							params.Fields[t] = append(params.Fields[t], f)
						}
					}
				}
			}
			// Check for duplicates
			for i := range params.Fields[t] {
				for j := i + 1; j < len(params.Fields[t]); j++ {
					if params.Fields[t][i] == params.Fields[t][j] {
						return nil, NewErrDuplicateFieldInFieldsParameter(
							typ.Name,
							params.Fields[t][i],
						)
					}
				}
			}
		}
	}

	// Attrs and Rels
	for typeName, fields := range params.Fields {
		// This should always return a type since
		// it is checked earlier.
		typ := schema.GetType(typeName)

		params.Attrs[typeName] = make([]Attr, 0, len(typ.Attrs))
		params.Rels[typeName] = make([]Rel, 0, len(typ.Attrs))

		// Get Attrs and Rels for the requested fields
		for _, field := range typ.Fields() {
			for _, field2 := range fields {
				if field == field2 {
					if typ = schema.GetType(typeName); typ.Name != "" {
						if attr, ok := typ.Attrs[field]; ok {
							// Append to list of attributes
							params.Attrs[typeName] = append(
								params.Attrs[typeName],
								attr,
							)
						} else if rel, ok := typ.Rels[field]; ok {
							// Append to list of relationships
							params.Rels[typeName] = append(
								params.Rels[typeName],
								rel,
							)
						}
					}
				}
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
				if !ok {
					return nil, NewErrUnknownSortRelationship(st.Name, parts[i])
				}

				if !rel.ToOne {
					return nil, NewErrInvalidSortRelationship(st.Name, parts[i])
				}

				path = append(path, rel)
				st = schema.GetType(rel.ToType)
			}

			sr.Name = parts[len(parts)-1]
			if _, ok := st.Attrs[sr.Name]; !ok && sr.Name != "id" {
				return nil, NewErrUnknownSortField(st.Name, sr.Name)
			}

			// Removes all redundant relationship paths. For example, a path of [a->b,b->a,a->c]
			// would be reduced to [a-c]. Self-relationships like [a->a] are removed.
			// todo: research if this can be implemented more effectively
			// todo: find a way to test this without having to use dozens of url with sort params
			for i := len(path) - 1; i >= 0; i-- {
				for j := 0; j <= i; j++ {
					if path[j].FromType == path[i].ToType {
						path = append(path[:j], path[i+1:]...)
						i = j

						break
					}
				}
			}

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

	return params, nil
}

// A Params object represents all the query parameters from the URL.
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
}
