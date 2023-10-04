package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

// assert is a helper function for testing that checks if two values are deeply equal.
// If they are not equal, it logs an error message.
func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		// If the two values are not equal, log an error with a descriptive message.
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLastMarketTrades(t *testing.T) {
	// Create a new order book
	ob := NewOrderbook()
	price := 10000.0

	// Create a sell order and place it in the order book
	sellOrder := NewOrder(false, 10, 0)
	ob.PlaceLimitOrder(price, sellOrder)

	// Create a market order and place it in the order book, check for matches
	marketOrder := NewOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(marketOrder)
	assert(t, len(matches), 1)
	match := matches[0]

	// Check the resulting trade and its properties
	assert(t, len(ob.Trades), 1)
	trade := ob.Trades[0]
	assert(t, trade.Price, price)
	assert(t, trade.Bid, marketOrder.Bid)
	assert(t, trade.Size, match.SizeFilled)
}

func TestLimit(t *testing.T) {
	// Create a new limit and some buy orders
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5, 0)
	buyOrderB := NewOrder(true, 8, 0)
	buyOrderC := NewOrder(true, 10, 0)

	// Add buy orders to the limit
	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	// Delete one buy order from the limit and print the limit
	l.DeleteOrder(buyOrderB)

	fmt.Println(l)
}

func TestPlaceLimitOrder(t *testing.T) {
	// Create a new order book
	ob := NewOrderbook()

	// Create sell orders and place them in the order book
	sellOrderA := NewOrder(false, 10, 0)
	sellOrderB := NewOrder(false, 5, 0)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(9_000, sellOrderB)

	// Check if the orders and ask limits are correctly stored
	assert(t, len(ob.Orders), 2)
	assert(t, ob.Orders[sellOrderA.ID], sellOrderA)
	assert(t, ob.Orders[sellOrderB.ID], sellOrderB)
	assert(t, len(ob.asks), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	// Create a new order book
	ob := NewOrderbook()

	// Create a sell order and place it in the order book
	sellOrder := NewOrder(false, 20, 0)
	ob.PlaceLimitOrder(10_000, sellOrder)

	// Create a buy order and place it as a market order, check for matches
	buyOrder := NewOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(buyOrder)

	// Check the matches and order book state
	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	// Create a new order book
	ob := NewOrderbook()

	// Create buy orders with various sizes
	buyOrderA := NewOrder(true, 5, 0) // filled fully
	buyOrderB := NewOrder(true, 8, 0) // partially filled
	buyOrderC := NewOrder(true, 1, 0)
	buyOrderD := NewOrder(true, 1, 0)

	// Place the buy orders in the order book
	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	// Check the total bid volume
	assert(t, ob.BidTotalVolume(), 15.00)

	// Create a sell order and place it as a market order, check for matches
	sellOrder := NewOrder(false, 10, 0)
	matches := ob.PlaceMarketOrder(sellOrder)

	// Check the order book state after market order execution
	assert(t, ob.BidTotalVolume(), 5.00)
	assert(t, len(ob.bids), 2)
	assert(t, len(matches), 2)
}

func TestCancelOrderAsk(t *testing.T) {
	// Create a new order book
	ob := NewOrderbook()

	// Create a sell order and place it in the order book
	sellOrder := NewOrder(false, 4, 0)
	price := 10_000.0
	ob.PlaceLimitOrder(price, sellOrder)

	// Check the ask total volume
	assert(t, ob.AskTotalVolume(), 4.0)

	// Cancel the sell order and check the order book state
	ob.CancelOrder(sellOrder)
	assert(t, ob.AskTotalVolume(), 0.0)

	_, ok := ob.Orders[sellOrder.ID]
	assert(t, ok, false)

	_, ok = ob.AskLimits[price]
	assert(t, ok, false)
}

func TestCancelOrderBid(t *testing.T) {
	// Create a new order book
	ob := NewOrderbook()

	// Create a buy order and place it in the order book
	buyOrder := NewOrder(true, 4, 0)
	price := 10_000.0
	ob.PlaceLimitOrder(price, buyOrder)

	// Check the bid total volume
	assert(t, ob.BidTotalVolume(), 4.0)

	// Cancel the buy order and check the order book state
	ob.CancelOrder(buyOrder)
	assert(t, ob.BidTotalVolume(), 0.0)

	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)

	_, ok = ob.BidLimits[price]
	assert(t, ok, false)
}
