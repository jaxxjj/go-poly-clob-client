// Example: fetching market data (L0 — no authentication needed).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	polyclob "github.com/jaxxjj/go-poly-clob-client"
)

func main() {
	client := polyclob.New("https://clob.polymarket.com", 137)
	ctx := context.Background()

	// Health check
	ok, err := client.GetOk(ctx)
	if err != nil {
		log.Fatalf("API health check failed: %v", err)
	}
	fmt.Println("API status:", ok)

	// Fetch a market by condition ID
	market, err := client.GetMarket(ctx, "0xbd31dc8a20211944f6b70f31557f1001557b59905b7738480ca09bd4532f84af")
	if err != nil {
		log.Fatalf("GetMarket: %v", err)
	}

	var m map[string]interface{}
	_ = json.Unmarshal(market, &m)
	fmt.Printf("Market: %s\n", m["question"])

	// Get order book for a token
	// (Use a real token ID from the market's clobTokenIds)
	tokenID := "71321045679252212594626385532706912750332728571942532289631379312455583992563"
	ob, err := client.GetOrderBook(ctx, tokenID)
	if err != nil {
		log.Fatalf("GetOrderBook: %v", err)
	}
	fmt.Printf("Order book: %d bids, %d asks\n", len(ob.Bids), len(ob.Asks))
	if len(ob.Bids) > 0 {
		fmt.Printf("  Best bid: %s @ %s\n", ob.Bids[0].Size, ob.Bids[0].Price)
	}
	if len(ob.Asks) > 0 {
		fmt.Printf("  Best ask: %s @ %s\n", ob.Asks[0].Size, ob.Asks[0].Price)
	}

	// Get midpoint price
	mid, err := client.GetMidpoint(ctx, tokenID)
	if err != nil {
		log.Fatalf("GetMidpoint: %v", err)
	}
	fmt.Printf("Midpoint: %s\n", string(mid))
}
