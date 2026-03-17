// Example: creating and posting an order (L1 + L2 authentication).
package main

import (
	"context"
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

	// Step 1: Create L1 client
	client, err := polyclob.NewL1("https://clob.polymarket.com", 137, privateKey)
	if err != nil {
		log.Fatalf("NewL1: %v", err)
	}
	fmt.Println("Address:", client.GetAddress())

	// Step 2: Derive L2 API credentials
	creds, err := client.CreateOrDeriveAPICreds(ctx, 0)
	if err != nil {
		log.Fatalf("CreateOrDeriveAPICreds: %v", err)
	}
	client.SetAPICreds(*creds)
	fmt.Println("API Key:", creds.APIKey)

	// Step 3: Check balance
	balance, err := client.GetBalanceAllowance(ctx, model.BalanceAllowanceParams{
		AssetType:     model.AssetCollateral,
		SignatureType: -1, // use client default
	})
	if err != nil {
		log.Fatalf("GetBalanceAllowance: %v", err)
	}
	fmt.Printf("Balance: %s\n", string(balance))

	// Step 4: Create and post an order
	tokenID := "71321045679252212594626385532706912750332728571942532289631379312455583992563"

	resp, err := client.CreateAndPostOrder(ctx, model.OrderArgs{
		TokenID: tokenID,
		Price:   0.50,
		Size:    10,
		Side:    "BUY",
	}, nil) // nil = auto-resolve tick_size and neg_risk
	if err != nil {
		log.Fatalf("CreateAndPostOrder: %v", err)
	}
	fmt.Printf("Order response: %s\n", string(resp))

	// Step 5: Check open orders
	orders, err := client.GetOrders(ctx, nil)
	if err != nil {
		log.Fatalf("GetOrders: %v", err)
	}
	fmt.Printf("Open orders: %d\n", len(orders))
}
