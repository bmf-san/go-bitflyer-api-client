package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/bmf-san/go-bitflyer-api-client/client/auth"
	"github.com/bmf-san/go-bitflyer-api-client/client/http"
)

func main() {
	// Load credentials from environment variables or use defaults
	credentials := auth.APICredentials{
		APIKey:    getEnvOrDefault("BITFLYER_API_KEY", "your-api-key"),
		APISecret: getEnvOrDefault("BITFLYER_API_SECRET", "your-api-secret"),
	}

	// Show which credential source is being used
	if os.Getenv("BITFLYER_API_KEY") != "" {
		fmt.Println("Using API credentials from environment variables")
	} else {
		fmt.Println("Using default API credentials (set BITFLYER_API_KEY and BITFLYER_API_SECRET environment variables)")
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
			productCode := "N/A"
			marketType := "N/A"
			if market.ProductCode != nil {
				productCode = *market.ProductCode
			}
			if market.MarketType != nil {
				marketType = string(*market.MarketType)
			}
			fmt.Printf("- %s (%s)\n", productCode, marketType)
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
		health := "N/A"
		state := "N/A"
		if boardState.JSON200.Health != nil {
			health = string(*boardState.JSON200.Health)
		}
		if boardState.JSON200.State != nil {
			state = string(*boardState.JSON200.State)
		}
		fmt.Printf("Health: %s\n", health)
		fmt.Printf("State: %s\n", state)
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

// getEnvOrDefault returns the environment variable value or a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
