# go-coinbase-trade
## Go library for the Coinbase Advanced Trade REST API

Reference: https://docs.cloud.coinbase.com/advanced-trade-api/docs/rest-api-overview

With the move from Coinbase Pro, Coinbase's trading API has been merged with the API for Coinbase's wallet functionality. This library only supports the Advanced Trade API, which includes the following methods:

- `List Accounts` - Get a list of all accounts (wallets)
- `Get Account` - Get details for one account
- `Create Order` - Place a new order on the exchange
- `Cancel Orders` - Cancel one or more orders that have already been placed
- `List Orders` - Get a list of all orders that meet the specified criteria
- `List Fills` - Get a list of all fills (matches) that meet the specified criteria
- `Get Order` - Get details for one order
- `List Products` - Get a list of all products (trading pairs)
- `Get Product` - Get details for one product
- `Get Product Candles` - Get historical market data for one product
- `Get Market Trades` - Get the latest trades for one product

## Credentials

To use this library, you will need to initalize a client with the API key and secret provided by your Coinbase account. The `Host` and `Path` values only need to be provided if you want to use something other than the production server (e.g. sandbox testing, etc)

There are two ways to provide your credentials when creating a client:

### Environment Variables (Recommended)

Set the following environment variables in your OS:

- `COINBASE_KEY` - Your API key
- `COINBASE_SECRET` - Your API secret
- `COINBASE_HOST` (Optional) - Use a host other than `https://coinbase.com`
- `COINBASE_PATH` (Optional) - Use a path other than `/api/v3/brokerage`

You can now create a client within your project without needing to specify any parameters:

```
client := coinbasetrade.NewClient(nil)
```

### Pass credentials in code

You can also populate a configuration object containing your credentials and include it when creating your client:

```
config := coinbasetrade.ClientConfig{
  Key: [your api key],
  Secret: [your api secret],
  Host: [optional host],
  Path: [optional path],
}

client := coinbasetrade.NewClient(&config)
```

## Numbers

All numbers related to monetary value or volume of orders use the Shopspring `decimal` library. Although the Coinbase API deals in strings, `coinbase-trade` will automatically convert numerical values to and from `decimal.Decimal` objects for use within your project. This keeps floating-point arithmetic precise and easy to manage.

## Lists

Calling any of the above methods that have `List` in their name will return an object prepopulated with the first page of results. Every list object will have a `Next()` function which will return `true` as long as there is still data to be consumed. Call `NextPage()` to update the object with the next set of data. To consume all data, continue calling `NextPage()` until `Next()` returns false:

```
list := client.ListAccounts()
for ; list.Next(); list.NextPage() {
  for i, v := range list.Accounts {
    // This loop will interate through all accounts
  }
}
```

## Orders

Orders store the product (trading pair) and side (buy or sell) for the order. They then use an `OrderConfiguration` object to specify the remaining details of the order (purchase price, size, expiration date, etc). These details will determine what type of order it is: Market, Limit (GTC/GTD), or Stop Loss (GTC/GTD). Due to the way the Coinbase API is designed, it is better to call `GetOrderConfiguration()` or `SetOrderConfiguration()` on an `Order` object instead of trying to access the details directly through the `Order` object. 

```
order, _ := client.GetOrder(orderID)
config := order.GetOrderConfiguration()

switch config.Type {
  case coinbasetrade.LimitGTC:
    // This is a limit order with no expiration time
    price := config.LimitPrice
  case coinbasetrade.LimitGTD:
    // This is a limit order with an `EndTime` value
  ...
}
```

### Placing a new order

```
config := coinbasetrade.OrderConfiguration{
  LimitPrice: decimal.NewFromFloat(10000),
  BaseSize: decimal.NewFromFlaot(0.1),
}

// each order needs a unique id that we set, we'll use the current time for this example
clientorderid := time.Now().String()
product := "BTC-USD"

placedOrder, _, err := client.CreateOrder(clientid, product, coinbasetrade.Buy, config)
```

### Updating the status of an order

To refresh the status of `placedOrder`, you can either pass it to the client to be updated in place:

```
err = client.UpdateOrder(placedOrder)
```

Or retrieve the order again, using the order id:

```
updatedOrder, err := client.GetOrder(placedOrder.ID)
```

## More information

If any details are lacking in this documentation, please open a new issue and I will be happy to elaborate.