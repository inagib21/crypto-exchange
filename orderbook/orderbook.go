package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// * Match represents a trade match between an Ask and a Bid order.
type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

// * Order represents a trading order.
type Order struct {
	ID        int64
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

// * Orders is a slice of Order objects.
type Orders []*Order

// * Implementing sorting methods for Orders.
func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

// * NewOrder creates a new Order object.
func NewOrder(bid bool, size float64) *Order {
	return &Order{
		ID:        int64(rand.Intn(100000000)),
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// * String method to display Order information.
func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

// * IsFilled checks if an Order is completely filled.
func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

// * Limit represents a price level with associated orders.
type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

// * Limits is a slice of Limit objects.
type Limits []*Limit

// * ByBestAsk is a sorting method for Limits based on Ask prices.
type ByBestAsk struct{ Limits }

// * Implementing sorting methods for ByBestAsk.
func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

// * ByBestBid is a sorting method for Limits based on Bid prices.
type ByBestBid struct{ Limits }

// * Implementing sorting methods for ByBestBid.
func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

// * NewLimit creates a new Limit object.
func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

// * AddOrder adds an Order to a Limit.
func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

// * DeleteOrder removes an Order from a Limit.
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

// * Fill matches an Order with other Orders in a Limit.
func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)
	for _, order := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.IsFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

		if o.IsFilled() {
			break
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

// * fillOrder matches two Orders within a Limit.
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
	if a.Size > b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
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

// * Orderbook represents the collection of asks and bids.
type Orderbook struct {
	asks []*Limit
	bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Order
}

// * NewOrderbook creates a new Orderbook.
func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}

// * PlaceMarketOrder places a market order in the Orderbook.
func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.Bid {
		if o.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("not enough volume [size:%.2f] for market order [size:%.2f]", ob.AskTotalVolume(), o.Size))
		}
		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}

		}
	} else {
		if o.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("not enough volume [size:%.2f] for market order [size:%.2f]", ob.BidTotalVolume(), o.Size))
		}
		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)
			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}

		}

	}

	return matches
}

// * PlaceLimitOrder places a limit order in the Orderbook.
func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

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
		ob.Orders[o.ID] = o
		limit.AddOrder(o)
	}
}

// * clearLimit removes a limit from the Orderbook.
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
}

// * CancelOrder cancels an order from the Orderbook.
func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)
}

// * BidTotalVolume calculates the total volume of all bid orders.
func (ob *Orderbook) BidTotalVolume() float64 {
	totalVolume := 0.0
	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}
	return totalVolume
}

// * AskTotalVolume calculates the total volume of all ask orders.
func (ob *Orderbook) AskTotalVolume() float64 {
	totalVolume := 0.0
	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}
	return totalVolume
}

// * Asks returns the sorted list of ask orders.
func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

// * Bids returns the sorted list of bid orders.
func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}
