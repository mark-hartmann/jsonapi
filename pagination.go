package jsonapi

import (
	"strconv"
)

// Paginator represents a pagination strategy.
type Paginator interface {
	// Params returns a list of all page parameters relevant for the pagination strategy.
	Params() []string
	// Set sets the value for the key.
	Set(key, value string) error
	// Encode returns a deterministic, url-encoded string representation of the Paginator.
	Encode() string
}

// BasicPaginator is a simple page-based paginator.
type BasicPaginator struct {
	Number uint
	Size   uint
}

func (p *BasicPaginator) Params() []string {
	return []string{"number", "size"}
}

func (p *BasicPaginator) Set(key, value string) error {
	if key == "number" {
		num, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return NewErrInvalidPageNumberParameter(value)
		}
		p.Number = uint(num)
	} else if key == "size" {
		size, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return NewErrInvalidPageSizeParameter(value)
		}
		p.Size = uint(size)
	}

	return nil
}

func (p *BasicPaginator) Encode() string {
	q := ""
	if p.Number > 0 {
		q += "page%5Bnumber%5D=" + strconv.Itoa(int(p.Number))
	}
	if p.Size > 0 {
		if len(q) > 0 {
			q += "&"
		}
		q += "page%5Bsize%5D=" + strconv.Itoa(int(p.Size))
	}
	return q
}
