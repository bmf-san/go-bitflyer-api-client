package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bmf-san/go-bitflyer-api-client/client/websocket"
)

func main() {
	// Create context
	ctx := context.Background()

	// Create WebSocket client
	client, err := websocket.NewClient(ctx, "wss://ws.lightstream.bitflyer.com/json-rpc")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close(ctx)

	// Set up data handlers
	client.OnTicker(func(ticker websocket.TickerMessage) {
		fmt.Printf("Ticker received: %s %.0f/%.0f (%.1f)\n",
			ticker.ProductCode,
			ticker.BestBid,
			ticker.BestAsk,
			ticker.Ltp)
	})

	client.OnExecutions(func(execs websocket.ExecutionsMessage) {
		fmt.Printf("Executions received for %s: %d executions\n",
			execs.ProductCode,
			len(execs.Executions))
		for i, exec := range execs.Executions {
			if i >= 3 {
				fmt.Println("  ...more")
				break
			}
			fmt.Printf("  %s %.0f @ %.0f\n", exec.Side, exec.Size, exec.Price)
		}
	})

	client.OnBoard(func(board websocket.BoardMessage) {
		fmt.Printf("Board update for %s: mid=%.0f, bids=%d, asks=%d\n",
			board.ProductCode,
			board.Data.MidPrice,
			len(board.Data.Bids),
			len(board.Data.Asks))

		// Output board data details
		if len(board.Data.Bids) > 0 {
			fmt.Printf("  Top bid: %.0f @ %.0f\n", board.Data.Bids[0].Size, board.Data.Bids[0].Price)
		}
		if len(board.Data.Asks) > 0 {
			fmt.Printf("  Top ask: %.0f @ %.0f\n", board.Data.Asks[0].Size, board.Data.Asks[0].Price)
		}
	})

	client.OnBoardSnapshot(func(snapshot websocket.BoardSnapshotMessage) {
		fmt.Printf("Board snapshot for %s: mid=%.0f, bids=%d, asks=%d\n",
			snapshot.ProductCode,
			snapshot.Data.MidPrice,
			len(snapshot.Data.Bids),
			len(snapshot.Data.Asks))

		// Output snapshot details
		if len(snapshot.Data.Bids) > 0 {
			fmt.Printf("  Top bid: %.0f @ %.0f\n", snapshot.Data.Bids[0].Size, snapshot.Data.Bids[0].Price)
		}
		if len(snapshot.Data.Asks) > 0 {
			fmt.Printf("  Top ask: %.0f @ %.0f\n", snapshot.Data.Asks[0].Size, snapshot.Data.Asks[0].Price)
		}
	})

	client.OnOrderEvents(func(event websocket.OrderEventMessage) {
		fmt.Printf("Order event: %s [%s] %s %.0f @ %.0f\n",
			event.ProductCode,
			event.EventType,
			event.Side,
			event.Size,
			event.Price)
	})

	// Subscribe to channels
	productCode := "BTC_JPY"
	channels := []string{
		fmt.Sprintf("lightning_ticker_%s", productCode),
		fmt.Sprintf("lightning_executions_%s", productCode),
		fmt.Sprintf("lightning_board_snapshot_%s", productCode),
		fmt.Sprintf("lightning_board_%s", productCode),
	}

	for _, ch := range channels {
		if err := client.Subscribe(ctx, ch); err != nil {
			log.Printf("Failed to subscribe to %s: %v", ch, err)
		} else {
			log.Printf("Subscribed to %s", ch)
		}
	}

	// Authenticate and subscribe to private channels
	if apiKey := os.Getenv("BITFLYER_API_KEY"); apiKey != "" {
		apiSecret := os.Getenv("BITFLYER_API_SECRET")
		if err := client.Auth(ctx, apiKey, apiSecret); err != nil {
			log.Printf("Failed to authenticate: %v", err)
		} else {
			log.Println("Authentication successful")

			// Subscribe to private channels
			privateChannels := []string{
				"child_order_events",
				"parent_order_events",
			}
			for _, ch := range privateChannels {
				if err := client.Subscribe(ctx, ch); err != nil {
					log.Printf("Failed to subscribe to %s: %v", ch, err)
				} else {
					log.Printf("Subscribed to %s", ch)
				}
			}
		}
	}

	// Wait for signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Waiting for data... (Press Ctrl+C to quit)")
	<-sigCh

	log.Println("Shutting down...")

	// Cleanup: Unsubscribe from subscribed channels
	for _, ch := range channels {
		if err := client.Unsubscribe(ctx, ch); err != nil {
			log.Printf("Failed to unsubscribe from %s: %v", ch, err)
		}
	}

	// Wait a moment before exiting
	time.Sleep(time.Second)
}
