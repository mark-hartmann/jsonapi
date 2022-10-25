package jsonapi

import (
	"errors"
	"net/url"
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
	Filter       map[string][]string
	SortingRules []string
	Page         map[string]string
	Include      []string
	// Params contains all off-spec query parameters.
	Params map[string][]string
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
		Params:       map[string][]string{},
	}

	if u == nil {
		return sURL, errors.New("jsonapi: pointer to url.URL is nil")
	}

	sURL.Fragments = parseFragments(u.Path)
	sURL.Route = deduceRoute(sURL.Fragments)

	values := u.Query()
	for name := range values {
		switch {
		case strings.HasPrefix(name, "fields[") && strings.HasSuffix(name, "]") && len(name) > 8:
			resType := name[7 : len(name)-1]

			if len(values.Get(name)) > 0 {
				sURL.Fields[resType] = parseCommaList(values.Get(name))
			}
		case strings.HasPrefix(name, "page[") && strings.HasSuffix(name, "]") && len(name) > 6:
			if sURL.Page == nil {
				sURL.Page = map[string]string{}
			}

			nme := name[5 : len(name)-1]
			sURL.Page[nme] = values.Get(name)
		case name == "filter" || strings.HasPrefix(name, "filter["):
			if sURL.Filter == nil {
				sURL.Filter = map[string][]string{}
			}

			sURL.Filter[name] = append(sURL.Filter[name], values[name]...)
		case name == "sort":
			for _, rules := range values[name] {
				sURL.SortingRules = append(sURL.SortingRules, parseCommaList(rules)...)
			}
		case name == "include":
			for _, include := range values[name] {
				sURL.Include = append(sURL.Include, parseCommaList(include)...)
			}
		default:
			sURL.Params[name] = values[name]
		}
	}

	return sURL, nil
}

// Path returns the path only of the SimpleURL. It does not include any query
// parameters.
func (s *SimpleURL) Path() string {
	return strings.Join(s.Fragments, "/")
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
