package coinbasetrade

import (
	"errors"
	"strconv"
)

// Pagination values need to be extracted from some API replies, but we would like to keep these
// values from being exposed outside this library. This struct is used for umarshaling pagination
// data from API responses, and then each request struct saves this information internally, in
// unexported fields.
type Pagination struct {
	parent     interface{}
	parameters interface{}
	method     Method
	endpoint   string

	client *Client
	noNext bool
	end    bool
	// pagination with cursor
	cursor string

	// pagination without cursor (limit must be non-zero)
	limit  int
	offset int
}

func (p *Pagination) Next() bool {
	return !p.end
}

func (p *Pagination) NextPage() error {
	if p.noNext {
		p.end = true
		return nil
	}

	pg := struct {
		HasNext     bool   `json:"has_next"`
		Cursor      string `json:"cursor"`
		NumProducts int    `json:"num_products"` // only used by offset pagination
	}{}

	query, err := parametersToValues(p.parameters)
	if err != nil {
		return err
	}

	if p.cursor != "" {
		query.Add("cursor", p.cursor)
	} else if p.offset > 0 { // only used by offset pagination
		query.Add("offset", strconv.Itoa(p.offset))
	}

	if err := p.client.Request(p.method, p.endpoint, query, []byte{}, p.parent, &pg); err != nil {
		return err
	}

	p.noNext, p.cursor = !pg.HasNext, pg.Cursor

	// if using offset pagination
	if p.cursor == "" {
		if p.limit == 0 {
			return errors.New("no limit specified for offset pagination")
		}
		p.offset += p.limit
		p.noNext = p.offset >= pg.NumProducts
	}
	return nil
}
