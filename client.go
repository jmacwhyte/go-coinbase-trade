package coinbasetrade

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Method string

const (
	apiInterval = time.Millisecond * 50 // the minimum amount of time to wait in between API calls
	apiTimeout  = time.Second * 60      // how long to wait for a response

	Get    Method = "GET"
	Put    Method = "PUT"
	Post   Method = "POST"
	Delete Method = "DELETE"

	listAccountsEndpoint          = "/accounts"
	getAccountEndpoint            = "/accounts/%s"
	createOrderEndpoint           = "/orders"
	cancelOrdersEndpoint          = "/orders/batch_cancel"
	listOrdersEndpoint            = "/orders/historical/batch"
	listFillsEndpoint             = "/orders/historical/fills"
	getOrderEndpoint              = "/orders/historical/%s"
	listProductsEndpoint          = "/products"
	getProductEndpoint            = "/products/%s"
	getProductCandlesEndpoint     = "/products/%s/candles"
	getMarketTradesEndpoint       = "/products/%s/ticker"
	getTransactionSummaryEndpoint = "/transaction_summary"
)

type Client struct {
	Host     string // i.e. coinbase.com
	Path     string // path to the api
	Key      string // API key as provided by Coinbase
	Secret   string // API secret as provided by Coinbase
	lastCall time.Time
	client   *http.Client

	debug bool
}

type ClientConfig struct {
	Host   string
	Path   string
	Key    string
	Secret string
}

func NewClient(config *ClientConfig) *Client {
	cc := Client{}
	if config != nil {
		cc = Client{
			Host:   config.Host,
			Path:   config.Path,
			Key:    config.Key,
			Secret: config.Secret,
		}
	}

	defaults := Client{
		Host: "https://coinbase.com",
		Path: "/api/v3/brokerage",
	}

	c := &Client{
		Host:   os.Getenv("COINBASE_HOST"),
		Path:   os.Getenv("COINBASE_PATH"),
		Key:    os.Getenv("COINBASE_KEY"),
		Secret: os.Getenv("COINBASE_SECRET"),
	}

	for _, v := range []Client{cc, defaults} {
		if c.Host == "" {
			c.Host = v.Host
		}
		if c.Path == "" {
			c.Path = v.Path
		}
		if c.Key == "" {
			c.Key = v.Key
		}
		if c.Secret == "" {
			c.Secret = v.Secret
		}
	}

	c.client = &http.Client{
		Timeout: apiTimeout,
	}
	c.lastCall = time.Now()
	return c
}

func (c *Client) Request(m Method, endpoint string, query url.Values, payload []byte, result, pagination interface{}) (err error) {

	// ensure we observe the minimum interval time
	time.Sleep(time.Until(c.lastCall.Add(apiInterval)))

	var data []byte
	var res *http.Response

	if data, res, err = c.request(m, endpoint, query, payload); err != nil {
		return
	}

	// if we don't get a success code
	if res.StatusCode != 200 {

		if c.debug {
			log.Printf("Error response: %s", data)
		}

		// attempt to unmarshal error
		e := struct {
			Message string `json:"message"`
		}{}
		if err = json.Unmarshal(data, &e); err != nil {
			// otherwise, return the body as the error
			e.Message = fmt.Sprintf("(%d) %s", res.StatusCode, data)
		}

		// if the api key or secret is missing, include that info to help debug
		if c.Key == "" || c.Secret == "" {
			e.Message += " [API key or secret is missing]"
		}

		return formatError("api response", errors.New(e.Message))
	}

	// if an interface was passed, try to unmarshal the response
	if result != nil {
		if err = json.Unmarshal(data, result); err != nil {
			if c.debug {
				log.Printf("API response causing error: %s\n", data)
			}

			return formatError("unmarshal api result", err)
		}
	}

	// if pagination data is requested, try to unmarshal that too
	if pagination != nil {
		if err = json.Unmarshal(data, &pagination); err != nil {
			return formatError("unmarshal pagination result", err)
		}
	}

	return
}

func (c *Client) request(m Method, endpoint string, query url.Values, payload []byte) (body []byte, res *http.Response, err error) {
	uri := fmt.Sprintf("%s%s%s?%s", c.Host, c.Path, endpoint, query.Encode())
	bod := bytes.NewReader(payload)

	// start the request
	var req *http.Request
	if req, err = http.NewRequest(string(m), uri, bod); err != nil {
		err = formatError("http request", err)
		return
	}

	// add headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Go Coinbase AT 1.0")

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	resource := c.Path + endpoint

	var signature string
	if signature, err = c.sign(timestamp, m, resource, payload); err != nil {
		err = formatError("generate signature", err)
		return
	}

	req.Header.Add("CB-ACCESS-KEY", c.Key)
	req.Header.Add("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("CB-ACCESS-SIGN", signature)

	// get the response and update last call time
	c.lastCall = time.Now()
	if res, err = c.client.Do(req); err != nil {
		err = formatError("http response", err)
		return
	}
	defer func() {
		res.Body.Close()
		c.lastCall = time.Now()
	}()

	if body, err = ioutil.ReadAll(res.Body); err != nil {
		err = formatError("read response body", err)
		return
	}
	return
}

func (c *Client) sign(timestamp string, method Method, resource string, data []byte) (sig string, err error) {
	hash := hmac.New(sha256.New, []byte(c.Secret))

	message := fmt.Sprintf("%s%s%s%s", timestamp, method, resource, data)
	if _, err = hash.Write([]byte(message)); err != nil {
		return
	}
	sig = hex.EncodeToString(hash.Sum(nil))
	return
}

func formatError(location string, err error) error {
	return errors.New(location + ": " + err.Error())
}

// EnableDebug turns on some extra logging information
func (c *Client) EnableDebug() {
	c.debug = true
}
