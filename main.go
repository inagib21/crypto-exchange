package main

import (
	"math/rand"
	"time"

	"github.com/inagib21/crypto-exchange/client"
	"github.com/inagib21/crypto-exchange/mm"
	"github.com/inagib21/crypto-exchange/server"
)

func main() {
	// Start the server in a goroutine.
	go server.StartServer()
	time.Sleep(1 * time.Second)

	c := client.NewClient()

	// Configuration for the Market Maker.
	cfg := mm.Config{
		UserID:         8,
		OrderSize:      10,
		MinSpread:      20,
		MakeInterval:   1 * time.Second,
		SeedOffset:     40,
		ExchangeClient: c,
		PriceOffset:    10,
	}
	maker := mm.NewMakerMaker(cfg)

	// Start the Market Maker.
	maker.Start()

	time.Sleep(2 * time.Second)

	// Start the market order placer in a goroutine.
	go marketOrderPlacer(c)

	select {}
}

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(500 * time.Millisecond)

	for {
		// Generate a random integer.
		randint := rand.Intn(10)
		bid := true
		if randint < 7 {
			bid = false
		}

		// Create a market order with random bid/ask and size.
		order := client.PlaceOrderParams{
			UserID: 7,
			Bid:    bid,
			Size:   1,
		}

		// Place the market order using the client.
		_, err := c.PlaceMarketOrder(&order)
		if err != nil {
			panic(err)
		}

		<-ticker.C
	}
}
