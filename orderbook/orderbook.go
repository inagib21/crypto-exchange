package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Trade represents a trade that occurred in the order book.
type Trade struct {
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

// Match represents a matching pair of ask and bid orders in the order book.
type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

// Order represents an order in the order book.
type Order struct {
	ID        int64
	UserID    int64
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

// Orders is a slice of Order pointers.
type Orders []*Order

// Len returns the length of the Orders slice.
func (o Orders) Len() int { return len(o) }

// Swap swaps two elements in the Orders slice.
func (o Orders) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

// Less compares two elements in the Orders slice based on Timestamp.
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

// NewOrder creates a new Order with the given bid, size, and user ID.
func NewOrder(bid bool, size float64, userID int64) *Order {
	return &Order{
		UserID:    userID,
		ID:        int64(rand.Intn(10000000)),
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// String returns a string representation of the Order.
func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f] | [id: %d]", o.Size, o.ID)
}

// Type returns the type of the Order, which can be "BID" or "ASK".
func (o *Order) Type() string {
	if o.Bid {
		return "BID"
	}
	return "ASK"
}

// IsFilled checks if the Order has been completely filled.
func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

// Limit represents a price level in the order book with associated orders.
type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

// Limits is a slice of Limit pointers.
type Limits []*Limit

// ByBestAsk sorts Limits by ascending Price.
type ByBestAsk struct{ Limits }

// Len returns the length of the Limits slice.
func (a ByBestAsk) Len() int { return len(a.Limits) }

// Swap swaps two elements in the Limits slice.
func (a ByBestAsk) Swap(i, j int) { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }

// Less compares two elements in the Limits slice based on Price.
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

// ByBestBid sorts Limits by descending Price.
type ByBestBid struct{ Limits }

// Len returns the length of the Limits slice.
func (b ByBestBid) Len() int { return len(b.Limits) }

// Swap swaps two elements in the Limits slice.
func (b ByBestBid) Swap(i, j int) { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }

// Less compares two elements in the Limits slice based on Price.
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

// NewLimit creates a new Limit with the given price.
func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

// AddOrder adds an order to the Limit.
func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

// DeleteOrder removes an order from the Limit.
func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}

	o.Limit = nil
	l.TotalVolume -= o.Size

	sort.Sort(l.Orders)
}

// Fill matches an order in the Limit with another order, resulting in one or more matches.
func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.Orders {
		if o.IsFilled() {
			break
		}

		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.IsFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

// fillOrder matches two orders in the Limit and returns a Match.
func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}

	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

// Orderbook represents an order book with asks, bids, trades, and order management.
type Orderbook struct {
	asks []*Limit
	bids []*Limit

	Trades []*Trade

	mu        sync.RWMutex
	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Order
}

// NewOrderbook creates a new Orderbook instance.
func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		Trades:    []*Trade{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}

// PlaceMarketOrder places a market order in the order book and returns any matches.
func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	matches := []Match{}

	if o.Bid {
		if o.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.AskTotalVolume(), o.Size))
		}

		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(false, limit)
			}
		}
	} else {
		if o.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.BidTotalVolume(), o.Size))
		}

		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}
		}
	}

	for _, match := range matches {
		trade := &Trade{
			Price:     match.Price,
			Size:      match.SizeFilled,
			Timestamp: time.Now().UnixNano(),
			Bid:       o.Bid,
		}
		ob.Trades = append(ob.Trades, trade)
	}

	logrus.WithFields(logrus.Fields{
		"currentPrice": ob.Trades[len(ob.Trades)-1].Price,
	}).Info()

	return matches
}

// PlaceLimitOrder places a limit order in the order book.
func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	ob.mu.Lock()
	defer ob.mu.Unlock()

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)

		if o.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	logrus.WithFields(logrus.Fields{
		"price":  limit.Price,
		"type":   o.Type(),
		"size":   o.Size,
		"userID": o.UserID,
	}).Info("new limit order")

	ob.Orders[o.ID] = o
	limit.AddOrder(o)
}

// clearLimit clears a limit price level from the order book.
func (ob *Orderbook) clearLimit(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}
		}
	}

	fmt.Printf("clearing limit price level [%.2f]\n", l.Price)
}

// CancelOrder cancels an order in the order book.
func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)

	if len(limit.Orders) == 0 {
		ob.clearLimit(o.Bid, limit)
	}
}

// BidTotalVolume returns the total volume of all bid orders in the order book.
func (ob *Orderbook) BidTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}

	return totalVolume
}

// AskTotalVolume returns the total volume of all ask orders in the order book.
func (ob *Orderbook) AskTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}

	return totalVolume
}

// Asks returns the ask orders sorted by price.
func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

// Bids returns the bid orders sorted by price.
func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}
