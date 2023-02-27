package coinbasetrade

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

type (
	// for orders
	Side                   string
	OrderStatus            string
	TimeInForce            string
	TriggerStatus          string
	OrderType              string
	OrderConfigurationType string
	StopDirection          string
	CreateOrderError       string
	CancelOrderError       string

	// for fills
	TradeType          string
	LiquidityIndicator string
)

const (
	Buy         Side = "BUY"
	Sell        Side = "SELL"
	UnknownSide      = "UNKNOWN_ORDER_SIDE"

	Pending       OrderStatus = "PENDING"
	Open          OrderStatus = "OPEN"
	Filled        OrderStatus = "FILLED"
	Cancelled     OrderStatus = "CANCELLED"
	Expired       OrderStatus = "EXPIRED"
	Failed        OrderStatus = "FAILED"
	UnknownStatus OrderStatus = "UNKNOWN_ORDER_STATUS"

	GoodUntilDateTime  TimeInForce = "GOOD_UNTIL_DATE_TIME"
	GoodUntilCancelled TimeInForce = "GOOD_UNTIL_CANCELLED"
	ImmediateOrCancel  TimeInForce = "IMMEDIATE_OR_CANCEL"
	FillOrKill         TimeInForce = "FILL_OR_KILL"
	UnknownTimeInForce TimeInForce = "UNKNOWN_TIME_IN_FORCE"

	InvalidOrderType     TriggerStatus = "INVALID_ORDER_TYPE"
	StopPending          TriggerStatus = "STOP_PENDING"
	StopTriggered        TriggerStatus = "STOP_TRIGGERED"
	UnknownTriggerStatus TriggerStatus = "UNKNOWN_TRIGGER_STATUS"

	Market           OrderType = "MARKET"
	Limit            OrderType = "LIMIT"
	Stop             OrderType = "STOP"
	StopLimit        OrderType = "STOP_LIMIT"
	UnknownOrderType OrderType = "UNKNOWN_ORDER_TYPE"

	MarketIOC                 OrderConfigurationType = "market_market_ioc"
	LimitGTC                  OrderConfigurationType = "limit_limit_gtc"
	LimitGTD                  OrderConfigurationType = "limit_limit_gtd"
	StopLimitGTC              OrderConfigurationType = "stop_limit_stop_limit_gtc"
	StopLimitGTD              OrderConfigurationType = "stop_limit_stop_limit_gtd"
	UnknownOrderConfiguration OrderConfigurationType = "unknown_order_config_type"

	StopDirectionUp      StopDirection = "STOP_DIRECTION_STOP_UP"
	StopDirectionDown    StopDirection = "STOP_DIRECTION_STOP_DOWN"
	UnknownStopDirection StopDirection = "UNKNOWN_STOP_DIRECTION"

	UnknownFailureReason          CreateOrderError = "UNKNOWN_FAILURE_REASON"
	UnsupportedOrderConfiguration CreateOrderError = "UNSUPPORTED_ORDER_CONFIGURATION"
	InvalidSide                   CreateOrderError = "INVALID_SIDE"
	InvalidProductId              CreateOrderError = "INVALID_PRODUCT_ID"
	InvalidSizePrecision          CreateOrderError = "INVALID_SIZE_PRECISION"
	InvalidPricePrecision         CreateOrderError = "INVALID_PRICE_PRECISION"
	InsufficientFund              CreateOrderError = "INSUFFICIENT_FUND"
	InvalidLedgerBalance          CreateOrderError = "INVALID_LEDGER_BALANCE"
	OrderEntryDisabled            CreateOrderError = "ORDER_ENTRY_DISABLED"
	IneligiblePair                CreateOrderError = "INELIGIBLE_PAIR"
	InvalidLimitPricePostOnly     CreateOrderError = "INVALID_LIMIT_PRICE_POST_ONLY"
	InvalidLimitPrice             CreateOrderError = "INVALID_LIMIT_PRICE"
	InvalidNoLiquidity            CreateOrderError = "INVALID_NO_LIQUIDITY"
	InvalidRequest                CreateOrderError = "INVALID_REQUEST"
	CommanderRejectedNewOrder     CreateOrderError = "COMMANDER_REJECTED_NEW_ORDER"
	InsufficientFunds             CreateOrderError = "INSUFFICIENT_FUNDS"

	UnknownCancelFailureReason   CancelOrderError = "UNKNOWN_CANCEL_FAILURE_REASON"
	InvalidCancelRequest         CancelOrderError = "INVALID_CANCEL_REQUEST"
	UnknownCancelOrder           CancelOrderError = "UNKNOWN_CANCEL_ORDER"
	CommanderRejectedCancelOrder CancelOrderError = "COMMANDER_REJECTED_CANCEL_ORDER"
	DuplicateCancelRequest       CancelOrderError = "DUPLICATE_CANCEL_REQUEST"

	TradeFill       = "FILL"
	TradeReversal   = "REVERSAL"
	TradeCorrection = "CORRECTION"
	TradeSynthetic  = "SYNTHETIC"

	LiquidityMaker   = "MAKER"
	LiquidityTaker   = "TAKER"
	LiquidityUnknown = "UNKNOWN_LIQUIDITY_INDICATOR"
)

// Order represents the status of an order that has been placed.
// NOTE: As of 12/2022, "reject reason" doesn't seem to have a very obvious use, so
// it is left as a string for now.
type Order struct {
	// used by ListOrders
	ID                 string             `json:"order_id,omitempty"`
	Product            string             `json:"product_id"`
	UserID             string             `json:"user_id,omitempty"`
	OrderConfiguration OrderConfiguration `json:"-"`
	// OrderConfiguration   OrderConfiguration `json:"order_configuration"`
	Side                 Side            `json:"side"`
	ClientOrderID        string          `json:"client_order_id"`
	Status               string          `json:"status,omitempty"`
	TimeInForce          TimeInForce     `json:"time_in_force,omitempty"`
	CreatedTime          time.Time       `json:"created_time,omitempty"`
	CompletionPercentage decimal.Decimal `json:"completion_percentage,omitempty"`
	FilledSize           decimal.Decimal `json:"filled_size,omitempty"`
	AverageFilledPrice   decimal.Decimal `json:"average_filled_price,omitempty"`
	Fee                  string          `json:"fee,omitempty"`
	NumberOfFills        decimal.Decimal `json:"number_of_fills,omitempty"`
	FilledValue          decimal.Decimal `json:"filled_value,omitempty"`
	PendingCancel        bool            `json:"pending_cancel,omitempty"`
	SizeInQuote          bool            `json:"size_in_quote,omitempty"`
	TotalFees            decimal.Decimal `json:"total_fees,omitempty"`
	SizeInclusiveOfFees  bool            `json:"size_inclusive_of_fees,omitempty"`
	TotalValueAfterFees  decimal.Decimal `json:"total_value_after_fees,omitempty"`
	TriggerStatus        TriggerStatus   `json:"trigger_status,omitempty"`
	Type                 OrderType       `json:"order_type,omitempty"`
	RejectReason         string          `json:"reject_reason,omitempty"`
	Settled              bool            `json:"settled,omitempty"`
	ProductType          ProductType     `json:"product_type,omitempty"`
	OutstandingHold      decimal.Decimal `json:"outstanding_hold_amount"`

	// used by GetOrder
	RejectMessage string `json:"reject_message,omitempty"`
	CancelMessage string `json:"cancel_message,omitempty"`
}

// OrderConfiguration includes all the possible settings for all order types. Due to how the API
// works, only one value is added to the OrderConfiguration map in the Order struct above, and the key
// is set to the type of order. Use GetOrderConfiguration and SetOrderConfiguration instead of accesing
// the map directly.
type OrderConfiguration struct {
	Type          OrderConfigurationType `json:"-"`
	QuoteSize     decimal.Decimal        `json:"quote_size,omitempty"`
	BaseSize      decimal.Decimal        `json:"base_size,omitempty"`
	LimitPrice    decimal.Decimal        `json:"limit_price,omitempty"`
	StopPrice     decimal.Decimal        `json:"stop_price,omitempty"`
	StopDirection StopDirection          `json:"stop_direction,omitempty"`
	EndTime       time.Time              `json:"-"`
	PostOnly      bool                   `json:"post_only,omitempty"`
}

// toMap builds a map of strings from the order config for use with the api
func (oc OrderConfiguration) toMap() (m map[string]string) {
	m = make(map[string]string)
	if !oc.QuoteSize.IsZero() {
		m["quote_size"] = oc.QuoteSize.String()
	}
	if !oc.BaseSize.IsZero() {
		m["base_size"] = oc.BaseSize.String()
	}
	if !oc.LimitPrice.IsZero() {
		m["limit_price"] = oc.LimitPrice.String()
	}
	if !oc.StopPrice.IsZero() {
		m["stop_price"] = oc.StopPrice.String()
	}
	if oc.StopDirection != "" {
		m["stop_direction"] = string(oc.StopDirection)
	}
	if !oc.EndTime.IsZero() {
		m["end_time"] = timeToString(oc.EndTime)
	}
	if oc.PostOnly {
		m["post_only"] = "true"
	}
	return
}

// getType returns the order configuration type, based on the values that are set
func (oc OrderConfiguration) getType() OrderConfigurationType {
	// classify order config
	gtd := !oc.EndTime.IsZero()
	stop := !oc.StopPrice.IsZero()
	limit := !oc.LimitPrice.IsZero()

	switch {
	case !limit: // if no limit price, it's a market order
		return MarketIOC
	case !gtd && !stop: // if no end date or stop price, it's a limit gtc
		return LimitGTC
	case gtd && !stop: // if there is an end date but no stop price, it's a limit gtd
		return LimitGTD
	case !gtd && stop: // if there is a stop price but no end date, it's a stop limit gtc
		return StopLimitGTC
	default: // must be a stop limit gtd
		return StopLimitGTD
	}
}

// CreateOrder will submit your raw order details and return a populated `Order` object. You must include a valid
// `OrderConfiguration` based on the type of order you wish to place. If the combination of data populated in
// the order config is invalid, the server will return an error. It is recommended to use one of the helper functions
// instead (PlaceMarketIOC, PlaceLimitGTC, etc)
func (c *Client) CreateOrder(clientOrderId string, productId string, side Side, orderConfig OrderConfiguration) (order Order, errorType CreateOrderError, err error) {

	// if no client id is specified, use unix time in milliseconds
	if clientOrderId == "" {
		clientOrderId = fmt.Sprintf("%d", time.Now().UnixMilli())
	}

	wrapper := struct {
		ClientOrderID      string                       `json:"client_order_id"`
		ProductID          string                       `json:"product_id"`
		Side               Side                         `json:"side"`
		OrderConfiguration map[string]map[string]string `json:"order_configuration"`
	}{clientOrderId, productId, side, map[string]map[string]string{string(orderConfig.Type): orderConfig.toMap()}}

	var payload []byte
	if payload, err = json.Marshal(wrapper); err != nil {
		err = formatError("create order", err)
		return
	}

	response := struct {
		Success     bool                          `json:"success"`
		OrderID     string                        `json:"order_id"`
		OrderConfig map[string]OrderConfiguration `json:"order_configuration"`
		Error       struct {
			Error   CreateOrderError `json:"error"`
			Details string           `json:"error_details"`
		} `json:"error_response"`
	}{}

	if _, err = c.makeRequest(Post, createOrderEndpoint, url.Values{}, payload, &response, nil); err != nil {
		err = formatError("api connection error", err)
		return
	}

	if response.Success {
		order = Order{
			ID:                 response.OrderID,
			Side:               side,
			OrderConfiguration: response.OrderConfig[string(orderConfig.getType())],
		}
		return
	}

	errorType = response.Error.Error
	err = errors.New(response.Error.Details)
	return
}

// CancelOrders takes a slice of order ids to cancel, and returns a map of potential errors for each order id.
func (c *Client) CancelOrders(orderIds []string) (cancelErrors map[string]CancelOrderError, err error) {
	wrapper := struct {
		Orders []string `json:"order_ids"`
	}{orderIds}

	var payload []byte
	if payload, err = json.Marshal(wrapper); err != nil {
		err = formatError("cancel orders", err)
		return
	}

	response := struct {
		Results []struct {
			Success bool             `json:"success"`
			Error   CancelOrderError `json:"failure_reason"`
			ID      string           `json:"order_id"`
		} `json:"results"`
	}{}

	if _, err = c.makeRequest(Post, cancelOrdersEndpoint, url.Values{}, payload, &response, nil); err != nil {
		err = formatError("api connection error", err)
		return
	}

	cancelErrors = make(map[string]CancelOrderError)
	for _, v := range response.Results {
		if v.Success {
			continue
		}
		cancelErrors[v.ID] = v.Error
	}
	if len(cancelErrors) > 0 {
		err = errors.New("one or more orders were not cancelled successfully")
	}
	return
}

type OrderList struct {
	Orders []Order `json:"orders"`
	Pagination
}

type ListOrdersParameters struct {
	Product            string        `cbt:"product_id"`
	Type               OrderType     `cbt:"order_type"`
	Side               Side          `cbt:"order_side"`
	Status             []OrderStatus `cbt:"order_status"`
	StartDate          time.Time     `cbt:"start_date"`
	EndDate            time.Time     `cbt:"end_date"`
	UserNativeCurrency string        `cbt:"user_native_currency"`
	ProductType        string        `cbt:"product_type"`
	Limit              int           `cbt:"limit"`
}

// ListOrders returns a list of orders based on the parameters you include.
func (c *Client) ListOrders(params ListOrdersParameters) (l OrderList, err error) {
	// this endpoint has no default limit, so we must ensure there is one
	if params.Limit <= 0 {
		params.Limit = 50
	}

	l.Pagination = Pagination{
		client:     c,
		parent:     &l,
		parameters: params,

		method:   Get,
		endpoint: listOrdersEndpoint,
	}

	err = l.NextPage()
	return
}

type Fill struct {
	ID                 string             `json:"entry_id"`
	TradeID            string             `json:"trade_id"`
	OrderID            string             `json:"order_id"`
	TradeTime          time.Time          `json:"trade_time"`
	Type               TradeType          `json:"trade_type"`
	Price              decimal.Decimal    `json:"price"`
	Size               decimal.Decimal    `json:"size"`
	Commission         decimal.Decimal    `json:"commission"`
	ProductID          string             `json:"product_id"`
	SequenceTime       time.Time          `json:"sequence_timestamp"`
	LiquidityIndicator LiquidityIndicator `json:"liquidity_indicator"`
	SizeInQuote        bool               `json:"size_in_quote"`
	UserID             string             `json:"user_id"`
	Side               Side               `json:"side"`
}

type FillList struct {
	Fills []Fill
	Pagination
}

type ListFillsParameters struct {
	OrderID           string    `cbt:"order_id"`
	ProductID         string    `cbt:"product_id"`
	StartSequenceTime time.Time `cbt:"start_sequence_timestamp"`
	EndSequenceTime   time.Time `cbt:"end_sequence_timestamp"`
	Limit             int       `cbt:"limit"`
}

// ListFills returns a list of fills based on the parameters you include.
func (c *Client) ListFills(params ListFillsParameters) (l FillList, err error) {
	l.Pagination = Pagination{
		client:     c,
		parent:     &l,
		parameters: params,

		method:   Get,
		endpoint: listFillsEndpoint,
	}

	err = l.NextPage()
	return
}

// GetOrder takes the order id assigned by Coinbase and returns a populated `Order` object containing the
// latest details from the server.
func (c *Client) GetOrder(id string) (o Order, err error) {
	// get order
	var data []byte
	if data, err = c.makeRequest(Get, fmt.Sprintf(getOrderEndpoint, id), url.Values{}, []byte{}, nil, nil); err != nil {
		return
	}

	// unmarshal the response, but the order config won't match up
	wrapper := &struct {
		Order *Order `json:"order"`
	}{&o}

	if err = json.Unmarshal(data, wrapper); err != nil {
		return
	}

	// unmarshal just the order config
	ocwrapper := &struct {
		Order struct {
			Config map[string]OrderConfiguration `json:"order_configuration"`
		} `json:"order"`
	}{}

	if err = json.Unmarshal(data, ocwrapper); err != nil {
		return
	}

	for _, v := range ocwrapper.Order.Config {
		o.OrderConfiguration = v
		break
	}
	o.OrderConfiguration.Type = o.OrderConfiguration.getType()

	return
}

// UpdateOrder takes an existing `Order` object, and updates it with the latest details from the server.
func (c *Client) UpdateOrder(order *Order) (err error) {
	var neworder Order
	if neworder, err = c.GetOrder(order.ID); err != nil {
		return
	}

	*order = neworder
	return
}

// PlaceMarketIOC is a helper function to place a market "immediate or cancel" order.
func (c *Client) PlaceMarketIOC(clientOrderId string, productId string, side Side, size decimal.Decimal) (order Order, errorType CreateOrderError, err error) {
	oc := OrderConfiguration{
		Type: MarketIOC,
	}
	if side == Buy {
		oc.QuoteSize = size
	} else {
		oc.BaseSize = size
	}
	return c.CreateOrder(clientOrderId, productId, side, oc)
}

// PlaceLimitGTC is a helper function to place a limit "good till closed" order. If you want to place
// a "post only" order, set postOnly to true.
func (c *Client) PlaceLimitGTC(clientOrderId string, productId string, side Side, size decimal.Decimal, price decimal.Decimal, postOnly bool) (order Order, errorType CreateOrderError, err error) {
	oc := OrderConfiguration{
		Type:       LimitGTC,
		BaseSize:   size,
		LimitPrice: price,
		PostOnly:   postOnly,
	}

	return c.CreateOrder(clientOrderId, productId, side, oc)
}

// PlaceLimitGTD is a helper function to place a limit "good till date" order. If you want to place
// a "post only" order, set postOnly to true.
func (c *Client) PlaceLimitGTD(clientOrderId string, productId string, side Side, size decimal.Decimal, price decimal.Decimal, endTime time.Time, postOnly bool) (order Order, errorType CreateOrderError, err error) {
	oc := OrderConfiguration{
		Type:       LimitGTD,
		BaseSize:   size,
		LimitPrice: price,
		EndTime:    endTime,
		PostOnly:   postOnly,
	}

	return c.CreateOrder(clientOrderId, productId, side, oc)
}

// PlaceStopLimitGTC is a helper function to place a limit "good till close" order with a stop loss
// price.
func (c *Client) PlaceStopLimitGTC(clientOrderId string, productId string, side Side, size decimal.Decimal, price decimal.Decimal, stopPrice decimal.Decimal, stopDirection StopDirection) (order Order, errorType CreateOrderError, err error) {
	oc := OrderConfiguration{
		Type:          LimitGTD,
		BaseSize:      size,
		LimitPrice:    price,
		StopPrice:     stopPrice,
		StopDirection: stopDirection,
	}

	return c.CreateOrder(clientOrderId, productId, side, oc)
}

// PlaceStopLimitGTD is a helper function to place a limit "good till date" order with a stop loss
// price.
func (c *Client) PlaceStopLimitGTD(clientOrderId string, productId string, side Side, size decimal.Decimal, price decimal.Decimal, stopPrice decimal.Decimal, stopDirection StopDirection, endTime time.Time) (order Order, errorType CreateOrderError, err error) {
	oc := OrderConfiguration{
		Type:          LimitGTD,
		BaseSize:      size,
		LimitPrice:    price,
		StopPrice:     stopPrice,
		EndTime:       endTime,
		StopDirection: stopDirection,
	}

	return c.CreateOrder(clientOrderId, productId, side, oc)
}
