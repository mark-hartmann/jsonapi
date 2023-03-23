package jsonapi

import (
	"errors"
	"net/url"
	"strings"
)

// A SimpleURL represents a URL not validated nor supplemented from a schema, values
// are stored as they are found in the request url.
type SimpleURL struct {
	Fragments []string // [users, abc123, articles]
	Route     string   // /users/:id/articles

	// Fields contains all resource fields (attributes and relationships), grouped by
	// their resource type.
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
	sURL := SimpleURL{}

	if u == nil {
		return sURL, errors.New("jsonapi: pointer to url.URL is nil")
	}

	fragments := parseFragments(u.Path)
	if len(fragments) != 0 {
		sURL.Fragments = fragments
		sURL.Route = deduceRoute(sURL.Fragments)
	}

	suPage := map[string]string{}
	suFields := map[string][]string{}
	suFilter := map[string][]string{}
	suParams := map[string][]string{}

	var (
		suInclude, suSortingRules []string
	)

	values := u.Query()
	for name := range values {
		switch {
		case strings.HasPrefix(name, "fields[") && strings.HasSuffix(name, "]") &&
			len(name) > 8:
			resType := name[7 : len(name)-1]
			for _, fields := range values[name] {
				suFields[resType] = append(suFields[resType], parseCommaList(fields)...)
			}
		case strings.HasPrefix(name, "page[") && strings.HasSuffix(name, "]") &&
			len(name) > 6:
			nme := name[5 : len(name)-1]
			suPage[nme] = values.Get(name)
		case name == "filter" || strings.HasPrefix(name, "filter["):
			suFilter[name] = append(suFilter[name], values[name]...)
		case name == "sort":
			for _, rules := range values[name] {
				suSortingRules = append(suSortingRules, parseCommaList(rules)...)
			}
		case name == "include":
			for _, include := range values[name] {
				suInclude = append(suInclude, parseCommaList(include)...)
			}
		default:
			suParams[name] = values[name]
		}
	}

	if len(suFields) > 0 {
		sURL.Fields = suFields
	}

	if len(suFilter) > 0 {
		sURL.Filter = suFilter
	}

	if len(suPage) > 0 {
		sURL.Page = suPage
	}

	if len(suInclude) > 0 {
		sURL.Include = suInclude
	}

	if len(suParams) > 0 {
		sURL.Params = suParams
	}

	if len(suSortingRules) > 0 {
		sURL.SortingRules = suSortingRules
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
