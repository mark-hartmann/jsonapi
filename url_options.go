package jsonapi

import "strings"

// URLOptions provides configuration options for creating instances of SimpleURL and URL.
type URLOptions struct {
	// Prefix must be set if the path from which the actual API starts has a prefix, e.g. /api/resource/:id.
	Prefix string
	// Paginator is any pagination strategy, for example BasicPaginator. If nil, pagination is not supported for
	// this URL and creating it will return an error if page parameters are specified.
	Paginator Paginator
}

// Path takes a raw path and returns a new one ready to be consumed by the library according to the given options.
func (o *URLOptions) Path(p string) string {
	path := strings.TrimPrefix(strings.Trim(p, "/"), strings.Trim(o.Prefix, "/"))
	if strings.HasPrefix(path, "/") {
		return path
	} else {
		return "/" + path
	}
}
