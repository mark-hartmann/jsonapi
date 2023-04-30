package jsonapi

import (
	"fmt"
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
					return nil, &srcError{src: "include", error: &UnknownFieldError{
						Type:    typ.Name,
						Field:   word,
						asRel:   true,
						relPath: relPath(incs[i]),
					}}
				}

				params.Include[i][j] = incRel
			}
		}
	}

	// After these checks, only valid fields remain, representing either the resource ID or
	// one of the attributes or relations.
	for typeName, fields := range su.Fields {
		fields = removeDuplicates(fields)

		typ := schema.GetType(typeName)
		if typeName != resType && typ.Name == "" {
			return nil, &srcError{src: "fields", error: &UnknownTypeError{Type: typeName}}
		}

		// Check if the sparse fieldset contains any fields that does not exist on the type.
		if field := findFirstDifference(fields, typ.Fields()); field != "" && field != "id" {
			return nil, &srcError{src: "fields", error: &UnknownFieldError{
				Type:  typ.Name,
				Field: field,
			}}
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

		for _, field := range fields {
			if attr, ok := typ.Attrs[field]; ok {
				attrs = append(attrs, attr)
			} else if rel, ok := typ.Rels[field]; ok {
				rels = append(rels, rel)
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

	typ := schema.GetType(resType)

	// Check for conflicting sort fields.
	set := make(map[string]struct{}, len(su.SortingRules))

	for _, sr := range su.SortingRules {
		nsr := strings.TrimPrefix(sr, "-")
		if _, ok := set[nsr]; ok {
			c := sr
			if strings.HasPrefix(c, "-") {
				c = sr[1:]
			}

			return nil, &ConflictingValueError{param: "sort", value: sr, conflictValue: c}
		}

		set[nsr] = struct{}{}
	}

	for _, rule := range su.SortingRules {
		sr, err := ParseSortRule(schema, typ, rule)
		if err != nil {
			// ParseSortRule does not annotate the error with an error source because
			// it's just a helper function without referring to a document.
			return nil, &srcError{
				src:   "sort",
				error: fmt.Errorf("jsonapi: failed to parse sort rule: %w", err),
			}
		}

		params.SortRules = append(params.SortRules, sr)
	}

	// Filter
	if len(su.Filter) > 0 {
		params.Filter = make(map[string][]string, len(su.Filter))
		for n, f := range su.Filter {
			params.Filter[n] = make([]string, len(f))
			copy(params.Filter[n], f)
		}
	}

	// Pagination
	if len(su.Page) > 0 {
		params.Page = make(map[string]string, len(su.Page))
		for k, v := range su.Page {
			params.Page[k] = v
		}
	}

	// Off-Spec query params
	if len(su.Params) > 0 {
		params.Params = make(map[string][]string, len(su.Params))
		for n, p := range su.Params {
			params.Params[n] = make([]string, len(p))
			copy(params.Params[n], p)
		}
	}

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

	// Filter contains all filtering data. JSON:API is agnostic about the filtering strategy
	// used by a server, so the values must be validated independently.
	Filter map[string][]string

	// SortRules contains all sorting rules.
	SortRules []SortRule

	// Page contains all pagination data. JSON:API is agnostic about the pagination strategy
	// used by a server, so the values must be validated independently.
	Page map[string]string

	// Include contains cleaned up relationship paths.
	Include [][]Rel

	// Params contains all off-spec query parameters.
	Params map[string][]string
}
