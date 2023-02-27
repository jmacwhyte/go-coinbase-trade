package coinbasetrade

import (
	"fmt"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

type Account struct {
	ID               string    `json:"uuid"`
	Name             string    `json:"name"`
	Currency         string    `json:"currency"`
	AvailableBalance Balance   `json:"available_balance"`
	Default          bool      `json:"default"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DeletedAt        time.Time `json:"deleted_at"`
	Type             string    `json:"type"`
	Ready            bool      `json:"ready"`
	HoldBalance      Balance   `json:"hold"`
}

type Balance struct {
	Value    decimal.Decimal `json:"value"`
	Currency string          `json:"currency"`
}

type AccountList struct {
	Accounts []Account `json:"accounts"`
	Pagination
}

type ListAccountsParameters struct {
	Limit int `cbt:"limit"`
}

// ListAccounts takes parameters (ListAccountsParameters), and returns an AccountsList. The
// AccountsList starts out popualated with the first page of results, and the next page of
// results can be retrieved by calling NextPage(). Next() will show if there are more pages
// to be retrieved.
func (c *Client) ListAccounts(params ListAccountsParameters) (l AccountList, err error) {
	l.Pagination = Pagination{
		client:     c,
		parent:     &l,
		parameters: params,

		method:   Get,
		endpoint: listAccountsEndpoint,
	}

	err = l.NextPage()
	return
}

// GetAccount takes an account ID and returns an Account object.
func (c *Client) GetAccount(id string) (acc Account, err error) {
	wrapper := &struct {
		Account *Account `json:"account"`
	}{&acc}

	_, err = c.makeRequest(Get, fmt.Sprintf(getAccountEndpoint, id), url.Values{}, []byte{}, wrapper, nil)
	return
}
