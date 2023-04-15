package jsonapi

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// NewURL builds a URL from a SimpleURL and a schema for validating and
// supplementing the object with extra information.
func NewURL(schema *Schema, su SimpleURL) (*URL, error) {
	url := &URL{}

	// Route
	url.Route = su.Route

	// Fragments
	url.Fragments = make([]string, len(su.Fragments))
	copy(url.Fragments, su.Fragments)

	// IsCol, ResType, ResID, RelKind, Rel, BelongsToFilter
	var (
		typ Type
		ok  bool
	)

	if len(url.Fragments) == 0 {
		// todo: turn fragmentError into ErrInvalidURL
		return nil, &fragmentError{fmt.Errorf("jsonapi: empty path")}
	}

	if len(url.Fragments) >= 1 {
		if typ = schema.GetType(url.Fragments[0]); typ.Name == "" {
			return nil, &fragmentError{&UnknownTypeError{Type: url.Fragments[0]}}
		}

		if len(url.Fragments) == 1 {
			url.IsCol = true
			url.ResType = typ.Name
		}

		if len(url.Fragments) == 2 {
			url.IsCol = false
			url.ResType = typ.Name
			url.ResID = url.Fragments[1]
		}
	}

	if len(url.Fragments) >= 3 {
		relName := url.Fragments[len(url.Fragments)-1]
		if url.Rel, ok = typ.Rels[relName]; !ok {
			// No Parameter/Pointer because it's part of the url path.
			return nil, &fragmentError{&UnknownFieldError{
				Type:  typ.Name,
				Field: relName,
				asRel: true,
			}}
		}

		url.IsCol = !url.Rel.ToOne
		url.ResType = url.Rel.ToType
		url.BelongsToFilter = BelongsToFilter{
			Type:   url.Fragments[0],
			ID:     url.Fragments[1],
			Name:   url.Rel.FromName,
			ToName: url.Rel.ToName,
		}

		if len(url.Fragments) == 3 {
			url.RelKind = "related"
		} else if len(url.Fragments) == 4 {
			url.RelKind = "self"
		}
	}

	// Check if the request has invalid parameters before creating the params object.
	if !url.IsCol {
		switch {
		case len(su.SortingRules) > 0:
			return nil, &IllegalParameterError{Param: "sort", isResource: true}
		case len(su.Page) > 0:
			return nil, &IllegalParameterError{Param: "page", isResource: true}
		case len(su.Filter) > 0:
			return nil, &IllegalParameterError{Param: "filter", isResource: true}
		}
	}

	var err error
	if url.Params, err = NewParams(schema, su, url.ResType); err != nil {
		return nil, fmt.Errorf("jsonapi: failed to create jsonapi.Params: %w", err)
	}

	return url, nil
}

// NewURLFromRaw parses rawurl to make a *url.URL before making and returning a
// *URL.
func NewURLFromRaw(schema *Schema, rawurl string) (*URL, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, fmt.Errorf("jsonapi: failed to parse url.URL: %w", err)
	}

	// u is a valid errJSONPtr, error can be ignored.
	su, _ := NewSimpleURL(u)

	return NewURL(schema, su)
}

// A URL stores all the information from a URL formatted for a JSON:API request.
//
// The data structure allows to have more information than what the URL itself
// stores.
type URL struct {
	// URL
	Fragments []string // [users, u1, articles]
	Route     string   // /users/:id/articles

	// Data
	IsCol           bool
	ResType         string
	ResID           string
	RelKind         string
	Rel             Rel
	BelongsToFilter BelongsToFilter

	// Params
	Params *Params
}

// String returns a string representation of the URL where special characters
// are escaped.
//
// The URL is normalized, so it always returns exactly the same string given the
// same URL.
func (u *URL) String() string {
	// Path
	path := "/"
	for _, p := range u.Fragments {
		path += p + "/"
	}

	path = path[:len(path)-1]

	// Params
	urlParams := []string{}

	// Fields
	fields := make([]string, 0, len(u.Params.Fields))
	for key := range u.Params.Fields {
		fields = append(fields, key)
	}

	sort.Strings(fields)

	for _, typ := range fields {
		if len(u.Params.Fields[typ]) == 0 {
			continue
		}

		sort.Strings(u.Params.Fields[typ])

		param := "fields%5B" + typ + "%5D="
		for _, f := range u.Params.Fields[typ] {
			param += f + "%2C"
		}

		param = param[:len(param)-3]

		urlParams = append(urlParams, param)
	}

	// Inclusions
	if len(u.Params.Include) > 0 {
		param := "include="
		inclusions := make([]string, 0, len(u.Params.Include))

		for _, rels := range u.Params.Include {
			var r string
			for _, rel := range rels {
				r += rel.FromName + "."
			}

			inclusions = append(inclusions, r[:len(r)-1])
		}

		sort.Strings(inclusions)
		param += strings.Join(inclusions, "%2C")
		urlParams = append(urlParams, param)
	}

	// Filter
	if u.Params.Filter != nil {
		var filterParams []string

		for name, filters := range u.Params.Filter {
			for _, f := range filters {
				// name (e.g. filter[xyz]) may contain characters that need to be url
				// encoded as well.
				filterParams = append(filterParams, url.QueryEscape(name)+"="+url.QueryEscape(f))
			}
		}

		sort.Strings(filterParams)
		urlParams = append(urlParams, filterParams...)
	}

	// Pagination
	if u.IsCol {
		var pageParams []string
		for k, v := range u.Params.Page {
			pageParams = append(pageParams, "page%5B"+url.QueryEscape(k)+"%5D="+fmt.Sprint(v))
		}

		// Maps have no reliable order. One could also use sort.Slice(pageParams) with a function
		// that explicitly checks the actual names, e.g.:
		// params[i][:strings.Index(params[i], "=")] < params[i][:strings.Index(params[j], "=")]
		sort.Strings(pageParams)
		urlParams = append(urlParams, pageParams...)
	}

	// Sorting
	if len(u.Params.SortRules) > 0 {
		param := "sort="

		for _, sr := range u.Params.SortRules {
			rule := sr.Name

			if len(sr.Path) > 0 {
				for _, rel := range sr.Path {
					rule = rel.FromName + "." + rule
				}
			}

			if sr.Desc {
				rule = "-" + rule
			}

			param += rule + "%2C"
		}

		param = param[:len(param)-3]
		urlParams = append(urlParams, param)
	}

	params := "?"
	for _, param := range urlParams {
		params += param + "&"
	}

	params = params[:len(params)-1]

	return path + params
}

// UnescapedString returns the same thing as String, but special characters are
// not escaped.
func (u *URL) UnescapedString() string {
	str, _ := url.PathUnescape(u.String())
	// TODO Can an error occur?
	return str
}

// A BelongsToFilter represents a parent resource, used to filter out resources
// that are not children of the parent.
//
// For example, in /articles/abc123/comments, the parent is the article with the
// ID abc123.
type BelongsToFilter struct {
	Type   string
	ID     string
	Name   string
	ToName string
}
