# go-bitflyer-api-client

[![Go CI](https://github.com/bmf-san/go-bitflyer-api-client/actions/workflows/test.yml/badge.svg)](https://github.com/bmf-san/go-bitflyer-api-client/actions/workflows/test.yml)
[![GitHub license](https://img.shields.io/github/license/bmf-san/go-bitflyer-api-client)](https://github.com/bmf-san/go-bitflyer-api-client/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/bmf-san/go-bitflyer-api-client.svg)](https://github.com/bmf-san/go-bitflyer-api-client/releases)

bitFlyer Lightning API client for Go. Supports both REST API and WebSocket (realtime) API.

## Features

- Generated from OpenAPI and AsyncAPI specifications
- Type-safe API client with full type definitions
- Supports authentication for private APIs
- WebSocket client for realtime data
- Comprehensive test coverage
- Example code included

## Disclaimer

**This software is provided for informational and development purposes only and is not intended to constitute financial advice or investment decisions. The author assumes no responsibility for any loss or damage arising from the use of this software.**

**This library is not affiliated with bitFlyer in any way. Please review the terms of service of each API provider before use.**

**This library is provided "as is", without any warranties of accuracy, completeness, or future compatibility.**

## Installation

```bash
go get github.com/bmf-san/go-bitflyer-api-client
```

## Usage

### HTTP API (REST)

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/bmf-san/go-bitflyer-api-client/client/auth"
    "github.com/bmf-san/go-bitflyer-api-client/client/http"
)

func main() {
    // Create client (with optional API credentials for private APIs)
    credentials := auth.APICredentials{
        APIKey:    "your-api-key",    // Optional
        APISecret: "your-api-secret", // Optional
    }

    client, err := http.NewAuthenticatedClient(credentials, "")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get available markets
    markets, err := client.Client().GetV1GetmarketsWithResponse(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if markets.JSON200 != nil {
        for _, market := range *markets.JSON200 {
            fmt.Printf("Market: %s (%s)\n", market.ProductCode, market.MarketType)
        }
    }

    // Get board state
    boardState, err := client.Client().GetV1GetboardstateWithResponse(ctx, &http.GetV1GetboardstateParams{
        ProductCode: "BTC_JPY",
    })
    if err != nil {
        log.Fatal(err)
    }

    if boardState.JSON200 != nil {
        fmt.Printf("Health: %s\n", boardState.JSON200.Health)
        fmt.Printf("State: %s\n", boardState.JSON200.State)
    }
}
```

### WebSocket API (Realtime)

```go
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

    // Setup data handlers
    client.OnTicker(func(ticker websocket.TickerMessage) {
        fmt.Printf("Ticker: %s %.0f/%.0f (%.1f)\n",
            ticker.ProductCode,
            ticker.BestBid,
            ticker.BestAsk,
            ticker.Ltp)
    })

    client.OnExecutions(func(execs websocket.ExecutionsMessage) {
        fmt.Printf("Executions: %s - %d trades\n",
            execs.ProductCode, len(execs.Executions))
    })

    client.OnBoard(func(board websocket.BoardMessage) {
        fmt.Printf("Board update: %s - bids: %d, asks: %d\n",
            board.ProductCode,
            len(board.Data.Bids),
            len(board.Data.Asks))
    })

    client.OnBoardSnapshot(func(snapshot websocket.BoardSnapshotMessage) {
        fmt.Printf("Board snapshot: %s - bids: %d, asks: %d\n",
            snapshot.ProductCode,
            len(snapshot.Data.Bids),
            len(snapshot.Data.Asks))
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

    // Optional authentication for private channels
    if apiKey := os.Getenv("BITFLYER_API_KEY"); apiKey != "" {
        apiSecret := os.Getenv("BITFLYER_API_SECRET")
        if err := client.Auth(ctx, apiKey, apiSecret); err == nil {
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
        } else {
            log.Printf("Authentication failed: %v", err)
        }
    }

    // Wait for signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    log.Println("Waiting for data... (Press Ctrl+C to quit)")
    <-sigCh

    log.Println("Shutting down...")

    // Clean up: unsubscribe from channels
    for _, ch := range channels {
        if err := client.Unsubscribe(ctx, ch); err != nil {
            log.Printf("Failed to unsubscribe from %s: %v", ch, err)
        }
    }
}
```

## API Coverage

### HTTP API
- Public API
  - Market Information `/v1/getmarkets`
  - Order Book Information `/v1/getboard`
  - Order Book Status `/v1/getboardstate`
  - Ticker Information `/v1/getticker`
  - Execution History `/v1/getexecutions`
  - Chat `/v1/getchats`
  - Exchange Status `/v1/gethealth`
  - Currency Information `/v1/getcurrencies`
- Private API
  - Account Balance `/v1/me/getbalance`
  - Deposit Address `/v1/me/getaddresses`
  - Deposit History `/v1/me/getcoinins`
  - Deposit History (Fast) `/v1/me/getcoins`
  - Withdrawal History `/v1/me/getcoinouts`
  - Bank Account Information `/v1/me/getbankaccounts`
  - Withdrawal Request `/v1/me/withdraw`
  - Cancel Withdrawal Request `/v1/me/cancelchildorder`
  - Withdrawal Address Book `/v1/me/getwithdrawals`
  - Trade History `/v1/me/getexecutions`
  - Position List `/v1/me/getpositions`
  - Collateral Status `/v1/me/getcollateral`
  - Collateral History `/v1/me/getcollateralhistory`
  - Trading Commission `/v1/me/gettradingcommission`

### WebSocket API
- Public Channels
  - Ticker `lightning_ticker_*`
  - Executions `lightning_executions_*`
  - Order Book (Snapshot) `lightning_board_snapshot_*`
  - Order Book (Incremental) `lightning_board_*`
- Private Channels
  - Child Order Events `child_order_events`
  - Parent Order Events `parent_order_events`

## Development

### Prerequisites
- Go 1.25.0 or later
- make

### Quick Start
```bash
# Install development tools
make install-tools

# Generate code from specifications
make generate

# Run tests
make test

# Run linter
make lint

# Run examples
make example
```

## CI/CD
This project uses GitHub Actions for continuous integration:
- Automated testing and linting on push and pull requests
  - Go 1.25.x
  - golangci-lint v2.0.2
- Manual releases with version tags

# Contribution
We welcome issues and pull requests at any time.

Feel free to contribute!

Before contributing, please check the following documents:

- [CODE_OF_CONDUCT](https://github.com/bmf-san/go-bitflyer-api-client/blob/main/.github/CODE_OF_CONDUCT.md)
- [CONTRIBUTING](https://github.com/bmf-san/go-/blob/main/.github/CONTRIBUTING.md)

# Sponsors
If you like this project, consider sponsoring us!

[Github Sponsors - bmf-san](https://github.com/sponsors/bmf-san)

Alternatively, giving us a star would be appreciated!

It helps motivate us to continue maintaining this project. :D

# License
This project is licensed under the MIT License.

[LICENSE](https://github.com/bmf-san/go-bitflyer-api-client/blob/main/LICENSE)

