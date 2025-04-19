package websocket

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Client represents a bitFlyer WebSocket API client
type Client struct {
	conn                 *websocket.Conn
	wsURL                string
	mu                   sync.Mutex
	jsonRPCID            int
	subscribedChannels   map[string]struct{}
	messageHandlers      map[string]MessageHandler
	tickerHandler        func(TickerMessage)
	executionsHandler    func(ExecutionsMessage)
	boardHandler         func(BoardMessage)
	boardSnapshotHandler func(BoardSnapshotMessage)
	privateOrderHandler  func(OrderEventMessage)
}

// JSON-RPC message structure
type jsonRPCRequest struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// MessageHandler is a function type that processes messages from a specific channel
type MessageHandler func(channel string, data json.RawMessage)

// Message type definitions
type TickerMessage struct {
	ProductCode     string  `json:"product_code"`
	Timestamp       string  `json:"timestamp"`
	BestBid         float64 `json:"best_bid"`
	BestAsk         float64 `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64 `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64 `json:"total_ask_depth"`
	Ltp             float64 `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

type ExecutionsMessage struct {
	ProductCode string      `json:"product_code"`
	Executions  []Execution `json:"data"`
}

type Execution struct {
	ID                         int64   `json:"id"`
	Side                       string  `json:"side"`
	Price                      float64 `json:"price"`
	Size                       float64 `json:"size"`
	ExecDate                   string  `json:"exec_date"`
	BuyChildOrderAcceptanceID  string  `json:"buy_child_order_acceptance_id"`
	SellChildOrderAcceptanceID string  `json:"sell_child_order_acceptance_id"`
}

type BoardMessage struct {
	ProductCode string    `json:"product_code"`
	Data        BoardData `json:"data"`
}

type BoardSnapshotMessage struct {
	ProductCode string    `json:"product_code"`
	Data        BoardData `json:"data"`
}

type BoardData struct {
	MidPrice float64      `json:"mid_price"`
	Bids     []PriceLevel `json:"bids"`
	Asks     []PriceLevel `json:"asks"`
}

type PriceLevel struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

type OrderEventMessage struct {
	ProductCode            string  `json:"product_code"`
	ChildOrderID           string  `json:"child_order_id"`
	ChildOrderAcceptanceID string  `json:"child_order_acceptance_id"`
	EventType              string  `json:"event_type"`
	EventDate              string  `json:"event_date"`
	Side                   string  `json:"side"`
	Price                  float64 `json:"price"`
	Size                   float64 `json:"size"`
	ExpireDate             string  `json:"expire_date"`
	Reason                 string  `json:"reason"`
	ExecID                 int64   `json:"exec_id"`
	Commission             float64 `json:"commission"`
}

// NewClient creates a new WebSocket client
func NewClient(ctx context.Context, wsURL string) (*Client, error) {
	// Connect to WebSocket
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{})
	if err != nil {
		return nil, fmt.Errorf("websocket connection error: %w", err)
	}

	client := &Client{
		conn:               conn,
		wsURL:              wsURL,
		jsonRPCID:          1,
		subscribedChannels: make(map[string]struct{}),
		messageHandlers:    make(map[string]MessageHandler),
	}

	// Start message receiving loop
	go client.receiveMessages(ctx)

	return client, nil
}

// Close closes the WebSocket connection
func (c *Client) Close(ctx context.Context) {
	if c.conn != nil {
		err := c.conn.Close(websocket.StatusNormalClosure, "client closed")
		if err != nil {
			// Log the error or handle it according to context
			fmt.Printf("An error occurred while closing the WebSocket connection: %v\n", err)
		}
	}
}

// OnTicker sets a callback to receive ticker information
func (c *Client) OnTicker(handler func(TickerMessage)) {
	c.tickerHandler = handler
}

// OnExecutions sets a callback to receive execution information
func (c *Client) OnExecutions(handler func(ExecutionsMessage)) {
	c.executionsHandler = handler
}

// OnBoard sets a callback to receive order book information
func (c *Client) OnBoard(handler func(BoardMessage)) {
	c.boardHandler = handler
}

// OnBoardSnapshot sets a callback to receive order book snapshots
func (c *Client) OnBoardSnapshot(handler func(BoardSnapshotMessage)) {
	c.boardSnapshotHandler = handler
}

// OnOrderEvents sets a callback to receive order events
func (c *Client) OnOrderEvents(handler func(OrderEventMessage)) {
	c.privateOrderHandler = handler
}

// Auth authenticates for using private API
func (c *Client) Auth(ctx context.Context, apiKey, apiSecret string) error {
	// Get current timestamp (using Unix timestamp as int64)
	unixTime := time.Now().Unix()

	// Calculate signature
	h := hmac.New(sha256.New, []byte(apiSecret))
	timeStr := strconv.FormatInt(unixTime, 10)
	h.Write([]byte(timeStr))
	h.Write([]byte(apiKey))
	signature := hex.EncodeToString(h.Sum(nil))

	// Create authentication message
	authParams := map[string]interface{}{
		"api_key":   apiKey,
		"timestamp": unixTime,
		"nonce":     fmt.Sprintf("%d", c.getNextID()),
		"signature": signature,
	}

	// Send authentication message
	return c.sendJSONRPC(ctx, "auth", authParams)
}

// Subscribe subscribes to the specified channel
func (c *Client) Subscribe(ctx context.Context, channel string) error {
	c.mu.Lock()
	if _, exists := c.subscribedChannels[channel]; exists {
		c.mu.Unlock()
		return fmt.Errorf("channel %s already subscribed", channel)
	}
	c.subscribedChannels[channel] = struct{}{}
	c.mu.Unlock()

	params := map[string]string{
		"channel": channel,
	}

	return c.sendJSONRPC(ctx, "subscribe", params)
}

// Unsubscribe cancels subscription to the specified channel
func (c *Client) Unsubscribe(ctx context.Context, channel string) error {
	c.mu.Lock()
	if _, exists := c.subscribedChannels[channel]; !exists {
		c.mu.Unlock()
		return fmt.Errorf("channel %s not subscribed", channel)
	}
	delete(c.subscribedChannels, channel)
	c.mu.Unlock()

	params := map[string]string{
		"channel": channel,
	}

	return c.sendJSONRPC(ctx, "unsubscribe", params)
}

// getNextID gets the next JSON-RPC ID
func (c *Client) getNextID() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.jsonRPCID
	c.jsonRPCID++
	return id
}

// sendJSONRPC sends a JSON-RPC message
func (c *Client) sendJSONRPC(ctx context.Context, method string, params interface{}) error {
	request := jsonRPCRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.getNextID(),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := wsjson.Write(ctx, c.conn, request); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// receiveMessages is a continuous message receiving loop from WebSocket
func (c *Client) receiveMessages(ctx context.Context) {
	for {
		var response json.RawMessage
		err := wsjson.Read(ctx, c.conn, &response)
		if err != nil {
			// End if connection is closed
			return
		}

		// Process message (in background)
		go c.handleMessage(ctx, response)
	}
}

// handleMessage processes messages received from WebSocket
func (c *Client) handleMessage(ctx context.Context, rawMsg json.RawMessage) {
	var msg map[string]json.RawMessage
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return
	}

	// Check if parameters exist
	paramsRaw, ok := msg["params"]
	if !ok {
		return
	}

	// Parse parameters as a map
	var params map[string]json.RawMessage
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return
	}

	// Get channel name
	channelRaw, ok := params["channel"]
	if !ok {
		return
	}

	var channel string
	if err := json.Unmarshal(channelRaw, &channel); err != nil {
		return
	}

	// Call the appropriate handler based on the channel
	if strings.HasPrefix(channel, "lightning_ticker_") {
		if c.tickerHandler != nil && params["message"] != nil {
			var ticker TickerMessage
			if err := json.Unmarshal(params["message"], &ticker); err == nil {
				c.tickerHandler(ticker)
			}
		}
	} else if strings.HasPrefix(channel, "lightning_executions_") {
		if c.executionsHandler != nil && params["message"] != nil {
			var executions ExecutionsMessage
			if err := json.Unmarshal(params["message"], &executions); err == nil {
				c.executionsHandler(executions)
			}
		}
	} else if strings.HasPrefix(channel, "lightning_board_") && !strings.HasPrefix(channel, "lightning_board_snapshot_") {
		if c.boardHandler != nil && params["message"] != nil {
			// Extract product code from channel name (example: lightning_board_BTC_JPY -> BTC_JPY)
			productCode := strings.TrimPrefix(channel, "lightning_board_")

			// Parse BoardData directly
			var boardData BoardData
			if err := json.Unmarshal(params["message"], &boardData); err == nil {
				// Create BoardMessage
				board := BoardMessage{
					ProductCode: productCode,
					Data:        boardData,
				}
				c.boardHandler(board)
			}
		}
	} else if strings.HasPrefix(channel, "lightning_board_snapshot_") {
		if c.boardSnapshotHandler != nil && params["message"] != nil {
			// Extract product code from channel name (example: lightning_board_snapshot_BTC_JPY -> BTC_JPY)
			productCode := strings.TrimPrefix(channel, "lightning_board_snapshot_")

			// Parse BoardData directly
			var boardData BoardData
			if err := json.Unmarshal(params["message"], &boardData); err == nil {
				// Create BoardSnapshotMessage
				snapshot := BoardSnapshotMessage{
					ProductCode: productCode,
					Data:        boardData,
				}
				c.boardSnapshotHandler(snapshot)
			}
		}
	} else if channel == "child_order_events" || channel == "parent_order_events" {
		if c.privateOrderHandler != nil && params["message"] != nil {
			var event OrderEventMessage
			if err := json.Unmarshal(params["message"], &event); err == nil {
				c.privateOrderHandler(event)
			}
		}
	}

	// Call user-defined handler if it exists
	if handler, exists := c.messageHandlers[channel]; exists {
		handler(channel, paramsRaw)
	}
}
