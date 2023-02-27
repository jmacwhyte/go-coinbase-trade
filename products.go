package coinbasetrade

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type ProductType string

const (
	UnknownProductType ProductType = "UNKNOWN_PRODUCT_TYPE"
	ProductTypeSpot    ProductType = "SPOT"
)

type Granularity string

const (
	OneMinute     Granularity = "ONE_MINUTE"
	FiveMinute    Granularity = "FIVE_MINUTE"
	FifteenMinute Granularity = "FIFTEEN_MINUTE"
	ThirtyMinute  Granularity = "THIRTY_MINUTE"
	OneHour       Granularity = "ONE_HOUR"
	TwoHour       Granularity = "TWO_HOUR"
	SixHour       Granularity = "SIX_HOUR"
	OneDay        Granularity = "ONE_DAY"
)

type Product struct {
	ID                        string          `json:"product_id"`
	Price                     decimal.Decimal `json:"price"`
	Volume24h                 decimal.Decimal `json:"volume_24h"`
	PricePercentageChange24h  string          `json:"price_percentage_change_24h"`
	VolumePercentageChange24h string          `json:"volume_percentage_change_24h"`
	BaseIncrement             decimal.Decimal `json:"base_increment"`
	QuoteIncrement            decimal.Decimal `json:"quote_increment"`
	QuoteMinSize              decimal.Decimal `json:"quote_min_size"`
	QuoteMaxSize              decimal.Decimal `json:"quote_max_size"`
	BaseMinSize               decimal.Decimal `json:"base_min_size"`
	BaseMaxSize               decimal.Decimal `json:"base_max_size"`
	BaseName                  string          `json:"base_name"`
	QuoteName                 string          `json:"quote_name"`
	Watched                   bool            `json:"watched"`
	IsDisabled                bool            `json:"is_disabled"`
	New                       bool            `json:"new"`
	Status                    string          `json:"status"`
	CancelOnly                bool            `json:"cancel_only"`
	LimitOnly                 bool            `json:"limit_only"`
	PostOnly                  bool            `json:"post_only"`
	TradingDisabled           bool            `json:"trading_disabled"`
	AuctionMode               bool            `json:"auction_mode"`
	ProductType               string          `json:"product_type"`
	QuoteCurrencyID           string          `json:"quote_currency_id"`
	BaseCurrencyID            string          `json:"base_currency_id"`
	// currently appears to not be populated by CB:
	// MidMarketPrice            decimal.Decimal `json:"mid_market_price"`
}

type ProductList struct {
	Products []Product `json:"products"`
	Pagination
}

type ListProductsParameters struct {
	Limit int         `cbt:"limit"`
	Type  ProductType `cbt:"product_type"`
}

// ListProducts returns a list of products based on the parameters you provide.
func (c *Client) ListProducts(params ListProductsParameters) (l ProductList, err error) {
	if params.Limit <= 0 {
		params.Limit = 100
	}
	l.Pagination = Pagination{
		client:     c,
		parent:     &l,
		parameters: params,
		limit:      params.Limit,

		method:   Get,
		endpoint: listProductsEndpoint,
	}

	err = l.NextPage()
	return
}

// GetProduct takes a product ID and returns a Product object.
func (c *Client) GetProduct(id string) (prod Product, err error) {
	_, err = c.makeRequest(Get, fmt.Sprintf(getProductEndpoint, id), url.Values{}, []byte{}, &prod, nil)
	return
}

type Candle struct {
	StartString string `json:"start"`
	StartTime   time.Time
	StartUnix   int64

	Low    decimal.Decimal `json:"low"`
	High   decimal.Decimal `json:"high"`
	Open   decimal.Decimal `json:"open"`
	Close  decimal.Decimal `json:"close"`
	Volume decimal.Decimal `json:"volume"`
}

// GetProductCandles takes a product ID, start and end times for the period you want to see, and the granularity
// of data that should be returned.
// The start time for each interval is included in 3 formats for convenience: string, int64, and time.Time.
func (c *Client) GetProductCandles(id string, start, end time.Time, granularity Granularity) (candles []Candle, err error) {
	// wrapper for the api response
	var res struct {
		Candles []Candle `json:"candles"`
	}

	// this endpoint has 3 required query parameters, times must be UNIX timestamp
	query := make(url.Values)
	query.Add("start", fmt.Sprintf("%d", start.Unix()))
	query.Add("end", fmt.Sprintf("%d", end.Unix()))
	query.Add("granularity", string(granularity))

	_, err = c.makeRequest(Get, fmt.Sprintf(getProductCandlesEndpoint, id), query, []byte{}, &res, nil)
	candles = res.Candles

	for i, v := range candles {
		candles[i].StartUnix, _ = strconv.ParseInt(v.StartString, 10, 64)
		candles[i].StartTime = time.Unix(candles[i].StartUnix, 0)
	}

	return
}

type Trade struct {
	ID        string          `json:"trade_id"`
	ProductID string          `json:"product_id"`
	Price     decimal.Decimal `json:"price"`
	Size      decimal.Decimal `json:"size"`
	Time      time.Time       `json:"time"`
	Side      Side            `json:"side"`
	// As of February 2023, these are included in the api response but only contain empty values:
	// Bid       decimal.Decimal `json:"bid"`
	// Ask       decimal.Decimal `json:"ask"`
}

type MarketTrades struct {
	Trades  []Trade
	BestBid decimal.Decimal `json:"best_bid"`
	BestAsk decimal.Decimal `json:"best_ask"`
}

// GetMarketTrades will return the current best bid and ask, plus a slice of the last `n` trades
// from the ticker
func (c *Client) GetMarketTrades(product string, n int) (market MarketTrades, err error) {

	query := make(url.Values)
	query.Add("limit", fmt.Sprintf("%d", n))

	_, err = c.makeRequest(Get, fmt.Sprintf(getMarketTradesEndpoint, product), query, []byte{}, &market, nil)
	return
}
