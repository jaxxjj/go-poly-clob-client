// Example: cancelling orders and managing positions (L2 authentication).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	polyclob "github.com/jaxxjj/go-poly-clob-client"
	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
)

func main() {
	privateKey := os.Getenv("POLYMARKET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("set POLYMARKET_PRIVATE_KEY env var")
	}

	ctx := context.Background()

	// Create authenticated client
	client, err := polyclob.NewL1("https://clob.polymarket.com", 137, privateKey)
	if err != nil {
		log.Fatalf("NewL1: %v", err)
	}

	creds, err := client.CreateOrDeriveAPICreds(ctx, 0)
	if err != nil {
		log.Fatalf("Derive creds: %v", err)
	}
	client.SetAPICreds(*creds)

	// List open orders
	orders, err := client.GetOrders(ctx, nil)
	if err != nil {
		log.Fatalf("GetOrders: %v", err)
	}
	fmt.Printf("Open orders: %d\n", len(orders))

	// Cancel a specific order
	if len(orders) > 0 {
		var order map[string]interface{}
		_ = json.Unmarshal(orders[0], &order)
		orderID := order["id"].(string)

		resp, err := client.Cancel(ctx, orderID)
		if err != nil {
			log.Fatalf("Cancel: %v", err)
		}
		fmt.Printf("Cancelled order %s: %s\n", orderID, string(resp))
	}

	// Or cancel all orders at once
	resp, err := client.CancelAll(ctx)
	if err != nil {
		log.Fatalf("CancelAll: %v", err)
	}
	fmt.Printf("CancelAll: %s\n", string(resp))

	// Check trade history
	trades, err := client.GetTrades(ctx, &model.TradeParams{})
	if err != nil {
		log.Fatalf("GetTrades: %v", err)
	}
	fmt.Printf("Trades: %d\n", len(trades))
}
