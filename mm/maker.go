package mm

import (
	"time"

	"github.com/inagib21/crypto-exchange/client"
	"github.com/sirupsen/logrus"
)

// Config holds configuration parameters for the MarketMaker.
type Config struct {
	UserID         int64          // UserID is the identifier of the market maker.
	OrderSize      float64        // OrderSize is the size of orders placed by the market maker.
	MinSpread      float64        // MinSpread is the minimum desired spread between bid and ask prices.
	SeedOffset     float64        // SeedOffset is the offset used for seeding the market.
	ExchangeClient *client.Client // ExchangeClient is the client for interacting with the exchange.
	MakeInterval   time.Duration  // MakeInterval is the time interval between market maker actions.
	PriceOffset    float64        // PriceOffset is the offset applied to bid and ask prices.
}

// MarketMaker represents a market maker responsible for placing orders on the exchange.
type MarketMaker struct {
	userID         int64
	orderSize      float64
	minSpread      float64
	seedOffset     float64
	priceOffset    float64
	exchangeClient *client.Client
	makeInterval   time.Duration
}

// NewMakerMaker creates a new MarketMaker instance with the provided configuration.
func NewMakerMaker(cfg Config) *MarketMaker {
	return &MarketMaker{
		userID:         cfg.UserID,
		orderSize:      cfg.OrderSize,
		minSpread:      cfg.MinSpread,
		seedOffset:     cfg.SeedOffset,
		exchangeClient: cfg.ExchangeClient,
		makeInterval:   cfg.MakeInterval,
		priceOffset:    cfg.PriceOffset,
	}
}

// Start starts the MarketMaker and initiates the market making process.
func (mm *MarketMaker) Start() {
	logrus.WithFields(logrus.Fields{
		"id":           mm.userID,
		"orderSize":    mm.orderSize,
		"makeInterval": mm.makeInterval,
		"minSpread":    mm.minSpread,
		"priceOffset":  mm.priceOffset,
	}).Info("starting market maker")

	go mm.makerLoop()
}

// makerLoop is the main loop for the market maker.
func (mm *MarketMaker) makerLoop() {
	ticker := time.NewTicker(mm.makeInterval)

	for {
		bestBid, err := mm.exchangeClient.GetBestBid()
		if err != nil {
			logrus.Error(err)
			break
		}

		bestAsk, err := mm.exchangeClient.GetBestAsk()
		if err != nil {
			logrus.Error(err)
			break
		}

		if bestAsk.Price == 0 && bestBid.Price == 0 {
			if err := mm.seedMarket(); err != nil {
				logrus.Error(err)
				break
			}
			continue
		}

		if bestBid.Price == 0 {
			bestBid.Price = bestAsk.Price - mm.priceOffset*2
		}

		if bestAsk.Price == 0 {
			bestAsk.Price = bestBid.Price + mm.priceOffset*2
		}

		spread := bestAsk.Price - bestBid.Price

		if spread <= mm.minSpread {
			continue
		}

		if err := mm.placeOrder(true, bestBid.Price+mm.priceOffset); err != nil {
			logrus.Error(err)
			break
		}
		if err := mm.placeOrder(false, bestAsk.Price-mm.priceOffset); err != nil {
			logrus.Error(err)
			break
		}

		<-ticker.C
	}
}

// placeOrder places a limit order with the specified bid and price.
func (mm *MarketMaker) placeOrder(bid bool, price float64) error {
	bidOrder := &client.PlaceOrderParams{
		UserID: mm.userID,
		Size:   mm.orderSize,
		Bid:    bid,
		Price:  price,
	}
	_, err := mm.exchangeClient.PlaceLimitOrder(bidOrder)
	return err
}

// seedMarket seeds the market with initial bid and ask orders.
func (mm *MarketMaker) seedMarket() error {
	currentPrice := simulateFetchCurrentETHPrice()

	logrus.WithFields(logrus.Fields{
		"currentETHPrice": currentPrice,
		"seedOffset":      mm.seedOffset,
	}).Info("orderbooks empty => seeding market!")

	bidOrder := &client.PlaceOrderParams{
		UserID: mm.userID,
		Size:   mm.orderSize,
		Bid:    true,
		Price:  currentPrice - mm.seedOffset,
	}
	_, err := mm.exchangeClient.PlaceLimitOrder(bidOrder)
	if err != nil {
		return err
	}

	askOrder := &client.PlaceOrderParams{
		UserID: mm.userID,
		Size:   mm.orderSize,
		Bid:    false,
		Price:  currentPrice + mm.seedOffset,
	}
	_, err = mm.exchangeClient.PlaceLimitOrder(askOrder)

	return err
}

// simulateFetchCurrentETHPrice simulates fetching the current ETH price from another exchange.
func simulateFetchCurrentETHPrice() float64 {
	time.Sleep(80 * time.Millisecond)

	return 1000.0
}
