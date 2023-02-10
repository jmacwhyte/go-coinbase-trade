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
	ID                   string                        `json:"order_id,omitempty"`
	Product              string                        `json:"product_id"`
	UserID               string                        `json:"user_id,omitempty"`
	OrderConfiguration   map[string]OrderConfiguration `json:"order_configuration"`
	Side                 Side                          `json:"side"`
	ClientOrderID        string                        `json:"client_order_id"`
	Status               string                        `json:"status,omitempty"`
	TimeInForce          TimeInForce                   `json:"time_in_force,omitempty"`
	CreatedTime          time.Time                     `json:"created_time,omitempty"`
	CompletionPercentage decimal.Decimal               `json:"completion_percentage,omitempty"`
	FilledSize           decimal.Decimal               `json:"filled_size,omitempty"`
	AverageFilledPrice   decimal.Decimal               `json:"average_filled_price,omitempty"`
	Fee                  string                        `json:"fee,omitempty"`
	NumberOfFills        decimal.Decimal               `json:"number_of_fills,omitempty"`
	FilledValue          decimal.Decimal               `json:"filled_value,omitempty"`
	PendingCancel        bool                          `json:"pending_cancel,omitempty"`
	SizeInQuote          bool                          `json:"size_in_quote,omitempty"`
	TotalFees            decimal.Decimal               `json:"total_fees,omitempty"`
	SizeInclusiveOfFees  bool                          `json:"size_inclusive_of_fees,omitempty"`
	TotalValueAfterFees  decimal.Decimal               `json:"total_value_after_fees,omitempty"`
	TriggerStatus        TriggerStatus                 `json:"trigger_status,omitempty"`
	// This doesn't seem important:
	// Type                 OrderType                     `json:"order_type,omitempty"`
	RejectReason string      `json:"reject_reason,omitempty"`
	Settled      bool        `json:"settled,omitempty"`
	ProductType  ProductType `json:"product_type,omitempty"`

	// used by GetOrder
	RejectMessage string `json:"reject_message,omitempty"`
	CancelMessage string `json:"cancel_message,omitempty"`
}

// OrderConfiguration includes all the possible settings for all order types. Due to how the API
// works, only one value is added to the OrderConfiguration map in the Order struct above, and the key
// is set to the type of order. Use GetOrderConfiguration and SetOrderConfiguration instead of accesing
// the map directly.
type OrderConfiguration struct {
	Type          OrderConfigurationType
	QuoteSize     decimal.Decimal `json:"quote_size,omitempty"`
	BaseSize      decimal.Decimal `json:"base_size,omitempty"`
	LimitPrice    decimal.Decimal `json:"limit_price,omitempty"`
	StopPrice     decimal.Decimal `json:"stop_price,omitempty"`
	StopDirection StopDirection   `json:"stop_direction,omitempty"`
	EndTime       time.Time       `json:"end_time,omitempty"`
	PostOnly      bool            `json:"post_only,omitempty"`
}

// GetOrderConfiguration is the best way to retrieve the order configuration from an order that has
// been returned from the API
func (o *Order) GetOrderConfiguration() (c OrderConfiguration) {
	c.Type = UnknownOrderConfiguration
	for k, v := range o.OrderConfiguration {
		for _, t := range []OrderConfigurationType{
			MarketIOC,
			LimitGTC,
			LimitGTD,
			StopLimitGTC,
			StopLimitGTD,
		} {
			if k == string(t) {
				c = v
				c.Type = t
			}
		}
	}
	return
}

// SetOrderConfiguration is the best way to set the order configuration for an order you have yet to place.
func (o *Order) SetOrderConfiguration(c OrderConfiguration) {
	o.OrderConfiguration = map[string]OrderConfiguration{
		string(c.Type): c,
	}
}

// CreateOrder will submit your order details and return a populated `Order` object. You must include a valid
// `OrderConfiguration` based on the type of order you wish to place. If the combination of data populated in
// the order config is invalid, the server will return an error.
func (c *Client) CreateOrder(clientOrderId string, productId string, side Side, orderConfig OrderConfiguration) (order Order, errorType CreateOrderError, err error) {
	//classify order config
	gtd := !orderConfig.EndTime.IsZero()
	stop := !orderConfig.StopPrice.IsZero()
	limit := !orderConfig.LimitPrice.IsZero()

	var t OrderConfigurationType
	switch {
	case !limit:
		t = MarketIOC
	case !gtd && !stop:
		t = LimitGTC
	case gtd && !stop:
		t = LimitGTD
	case !gtd && stop:
		t = StopLimitGTC
	default:
		t = StopLimitGTD
	}

	wrapper := struct {
		ClientOrderID      string                        `json:"client_order_id"`
		ProductID          string                        `json:"product_id"`
		Side               Side                          `json:"side"`
		OrderConfiguration map[string]OrderConfiguration `json:"order_configuration"`
	}{clientOrderId, productId, side, map[string]OrderConfiguration{string(t): orderConfig}}

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

	if err = c.Request(Post, createOrderEndpoint, url.Values{}, payload, &response, nil); err != nil {
		err = formatError("api connection error", err)
		return
	}

	if response.Success {
		order = Order{
			ID:                 response.OrderID,
			Side:               side,
			OrderConfiguration: response.OrderConfig,
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

	if err = c.Request(Post, cancelOrdersEndpoint, url.Values{}, payload, &response, nil); err != nil {
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
	wrapper := &struct {
		Order *Order `json:"order"`
	}{&o}

	err = c.Request(Get, fmt.Sprintf(getOrderEndpoint, id), url.Values{}, []byte{}, wrapper, nil)
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
