package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bmf-san/go-bitflyer-api-client/client/auth"
	"github.com/bmf-san/go-bitflyer-api-client/client/http"
)

func main() {
	// Create credentials (replace with your actual API credentials)
	credentials := auth.APICredentials{
		APIKey:    "your-api-key",
		APISecret: "your-api-secret",
	}

	// Create authenticated client
	client, err := http.NewAuthenticatedClient(credentials, "")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get markets (public API)
	ctx := context.Background()
	markets, err := client.Client().GetV1GetmarketsWithResponse(ctx)
	if err != nil {
		log.Fatalf("Failed to get markets: %v", err)
	}

	if markets.JSON200 != nil {
		fmt.Println("Available markets:")
		for _, market := range *markets.JSON200 {
			fmt.Printf("- %v (%v)\n", market.ProductCode, market.MarketType)
		}
	}

	// Get board state for BTC_JPY (public API)
	boardState, err := client.Client().GetV1GetboardstateWithResponse(ctx, &http.GetV1GetboardstateParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		log.Fatalf("Failed to get board state: %v", err)
	}

	if boardState.JSON200 != nil {
		fmt.Printf("\nBTC_JPY Board State:\n")
		fmt.Printf("Health: %v\n", boardState.JSON200.Health)
		fmt.Printf("State: %v\n", boardState.JSON200.State)
	}

	// Get ticker for BTC_JPY (public API) - Using raw response
	resp, err := client.Client().GetV1Getticker(ctx, &http.GetV1GettickerParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		log.Printf("Failed to get ticker: %v", err)
	} else {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Failed to close response body: %v", err)
			}
		}()

		// Parse response as raw JSON
		var tickerRaw map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&tickerRaw); err != nil {
			log.Printf("Failed to decode ticker response: %v", err)
		} else {
			fmt.Printf("\nBTC_JPY Ticker:\n")
			if bid, ok := tickerRaw["best_bid"].(float64); ok {
				fmt.Printf("Best Bid: %v\n", bid)
			}
			if ask, ok := tickerRaw["best_ask"].(float64); ok {
				fmt.Printf("Best Ask: %v\n", ask)
			}
			if ltp, ok := tickerRaw["ltp"].(float64); ok {
				fmt.Printf("Last Price: %v\n", ltp)
			}
			if vol, ok := tickerRaw["volume"].(float64); ok {
				fmt.Printf("Volume: %v\n", vol)
			}
			if ts, ok := tickerRaw["timestamp"].(string); ok {
				fmt.Printf("Timestamp: %v\n", ts)
			}
		}
	}
}
