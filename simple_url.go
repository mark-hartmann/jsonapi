package jsonapi

import (
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// A SimpleURL represents a URL not validated nor supplemented from a schema.
//
// It parses a URL in text format and stores the values as is.
type SimpleURL struct {
	// Source string

	// URL
	Fragments []string // [users, abc123, articles]
	Route     string   // /users/:id/articles

	// Params
	Fields       map[string][]string
	FilterLabel  string
	Filter       *Filter
	SortingRules []string
	PageSize     uint
	PageNumber   uint
	Include      []string
}

// NewSimpleURL takes and parses a *url.URL and returns a SimpleURL.
func NewSimpleURL(u *url.URL) (SimpleURL, error) {
	sURL := SimpleURL{
		Fragments: []string{},
		Route:     "",

		Fields:       map[string][]string{},
		Filter:       nil,
		SortingRules: []string{},
		Include:      []string{},
	}

	if u == nil {
		return sURL, errors.New("jsonapi: pointer to url.URL is nil")
	}

	sURL.Fragments = parseFragments(u.Path)
	sURL.Route = deduceRoute(sURL.Fragments)

	values := u.Query()
	for name := range values {
		if strings.HasPrefix(name, "fields[") && strings.HasSuffix(name, "]") && len(name) > 8 {
			// Fields
			resType := name[7 : len(name)-1]

			if len(values.Get(name)) > 0 {
				sURL.Fields[resType] = parseCommaList(values.Get(name))
			}
		} else {
			switch name {
			case "filter":
				var err error
				if values.Get(name)[0] != '{' {
					// It should be a label
					err = json.Unmarshal([]byte("\""+values.Get(name)+"\""), &sURL.FilterLabel)
				} else {
					// It should be a JSON object
					sURL.Filter = &Filter{}
					err = json.Unmarshal([]byte(values.Get(name)), sURL.Filter)
				}

				if err != nil {
					sURL.FilterLabel = ""
					sURL.Filter = nil

					return sURL, NewErrMalformedFilterParameter(values.Get(name))
				}
			case "sort":
				// Sort
				for _, rules := range values[name] {
					sURL.SortingRules = append(sURL.SortingRules, parseCommaList(rules)...)
				}
			case "page[size]":
				// Page size
				size, err := strconv.ParseUint(values.Get(name), 10, 64)
				if err != nil {
					return sURL, NewErrInvalidPageSizeParameter(values.Get(name))
				}

				sURL.PageSize = uint(size)
			case "page[number]":
				// Page number
				num, err := strconv.ParseUint(values.Get(name), 10, 64)
				if err != nil {
					return sURL, NewErrInvalidPageNumberParameter(values.Get(name))
				}

				sURL.PageNumber = uint(num)
			case "include":
				// Include
				for _, include := range values[name] {
					sURL.Include = append(sURL.Include, parseCommaList(include)...)
				}
			default:
				// Unkmown parameter
				return sURL, NewErrUnknownParameter(name)
			}
		}
	}

	return sURL, nil
}

// Path returns the path only of the SimpleURL. It does not include any query
// parameters.
func (s *SimpleURL) Path() string {
	return strings.Join(s.Fragments, "/")
}

// String returns a string representation of the SimpleURL where special characters
// are escaped. The SimpleURL is normalized, so it always returns exactly the same
// string given the same SimpleURL.
func (s *SimpleURL) String() string {
	path := "/"
	for _, p := range s.Fragments {
		path += p + "/"
	}

	path = path[:len(path)-1]
	urlParams := []string{}

	// Includes
	if len(s.Include) > 0 {
		param := "include="
		for _, include := range s.Include {
			param += include + "%2C"
		}

		param = param[:len(param)-3]
		urlParams = append(urlParams, param)
	}

	// Fields
	fields := make([]string, 0, len(s.Fields))
	for key := range s.Fields {
		fields = append(fields, key)
	}

	sort.Strings(fields)

	for _, typ := range fields {
		sort.Strings(s.Fields[typ])
		param := "fields%5B" + typ + "%5D="

		for _, f := range s.Fields[typ] {
			param += f + "%2C"
		}

		param = param[:len(param)-3] // Remove the last %2C
		urlParams = append(urlParams, param)
	}

	// Filter.
	if s.Filter != nil {
		mf, err := json.Marshal(s.Filter)
		if err != nil {
			// This should not happen since Filter should be validated
			// at this point.
			panic(err)
		}

		param := "filter=" + string(mf)
		urlParams = append(urlParams, param)
	} else if s.FilterLabel != "" {
		urlParams = append(urlParams, "filter="+s.FilterLabel)
	}

	// Pagination
	if s.PageNumber != 0 {
		urlParams = append(urlParams, "page%5Bnumber%5D="+strconv.Itoa(int(s.PageNumber)))
	}

	if s.PageSize != 0 {
		urlParams = append(urlParams, "page%5Bsize%5D="+strconv.Itoa(int(s.PageSize)))
	}

	// Sorting
	if len(s.SortingRules) > 0 {
		param := "sort="
		for _, attr := range s.SortingRules {
			param += attr + "%2C"
		}

		param = param[:len(param)-3] // Remove the last %2C
		urlParams = append(urlParams, param)
	}

	params := "?"
	for _, param := range urlParams {
		params += param + "&"
	}

	params = params[:len(params)-1] // Remove the last &

	return path + params
}

func (s *SimpleURL) UnescapedString() string {
	str, _ := url.PathUnescape(s.String())
	return str
}

func parseCommaList(path string) []string {
	items := strings.Split(path, ",")
	items2 := make([]string, 0, len(items))

	for i := range items {
		if items[i] != "" {
			items2 = append(items2, items[i])
		}
	}

	return items2
}

func parseFragments(path string) []string {
	fragments := strings.Split(path, "/")
	fragments2 := make([]string, 0, len(fragments))

	for i := range fragments {
		if fragments[i] != "" {
			fragments2 = append(fragments2, fragments[i])
		}
	}

	return fragments2
}

func deduceRoute(path []string) string {
	const (
		id   = "/:id"
		meta = "meta"
		rel  = "relationships"
	)

	route := ""

	if len(path) >= 1 {
		route = "/" + path[0]
	}

	if len(path) >= 2 {
		if path[1] == meta {
			route += "/" + meta
		} else {
			route += id
		}
	}

	if len(path) >= 3 {
		switch {
		case path[2] == rel:
			route += "/" + rel
		case path[2] == meta:
			route += "/" + meta
		default:
			route += "/" + path[2]
		}
	}

	if len(path) >= 4 {
		if path[3] == meta {
			route += "/" + meta
		} else if path[2] == rel {
			route += "/" + path[3]
		}
	}

	if len(path) >= 5 {
		if path[4] == meta {
			route += "/" + meta
		}
	}

	return route
}
