package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-bitflyer-api-client/client/auth"
)

// Helper functions
func str(s string) *string {
	return &s
}

func i(i int) *int {
	return &i
}

func f32(f float32) *float32 {
	return &f
}

func marketType(s string) *MarketMarketType {
	mt := MarketMarketType(s)
	return &mt
}

func boardStateHealth(s string) *BoardStateHealth {
	h := BoardStateHealth(s)
	return &h
}

func boardStateState(s string) *BoardStateState {
	s2 := BoardStateState(s)
	return &s2
}

func conditionType(s string) ParentOrderParameterConditionType {
	return ParentOrderParameterConditionType(s)
}

func side(s string) ParentOrderParameterSide {
	return ParentOrderParameterSide(s)
}

func orderMethod(s string) *NewParentOrderRequestOrderMethod {
	m := NewParentOrderRequestOrderMethod(s)
	return &m
}

// Public API tests
func TestGetMarkets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getmarkets" {
			t.Errorf("Expected path /v1/getmarkets, got %s", r.URL.Path)
		}

		markets := []Market{
			{
				ProductCode: str("BTC_JPY"),
				MarketType:  marketType("Spot"),
			},
			{
				ProductCode: str("FX_BTC_JPY"),
				MarketType:  marketType("FX"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(markets); err != nil {
			t.Fatalf("Failed to encode markets: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetmarketsWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get markets: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected markets response, got nil")
	}

	if len(*resp.JSON200) != 2 {
		t.Errorf("Expected 2 markets, got %d", len(*resp.JSON200))
	}

	market := (*resp.JSON200)[0]
	if *market.ProductCode != "BTC_JPY" {
		t.Errorf("Expected BTC_JPY, got %s", *market.ProductCode)
	}
	if *market.MarketType != "Spot" {
		t.Errorf("Expected Spot market type, got %s", *market.MarketType)
	}
}

func TestGetBoard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getboard" {
			t.Errorf("Expected path /v1/getboard, got %s", r.URL.Path)
		}

		entries := []BoardEntry{
			{
				Price: f32(2999000),
				Size:  f32(0.1),
			},
		}

		askEntries := []BoardEntry{
			{
				Price: f32(3001000),
				Size:  f32(0.2),
			},
		}

		board := Board{
			MidPrice: f32(3000000),
			Bids:     &entries,
			Asks:     &askEntries,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(board); err != nil {
			t.Fatalf("Failed to encode board: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetboardWithResponse(ctx, &GetV1GetboardParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get board: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected board response, got nil")
	}

	board := resp.JSON200
	if *board.MidPrice != 3000000 {
		t.Errorf("Expected mid price 3000000, got %f", *board.MidPrice)
	}

	if len(*board.Bids) != 1 {
		t.Errorf("Expected 1 bid, got %d", len(*board.Bids))
	}
	if *(*board.Bids)[0].Price != 2999000 {
		t.Errorf("Expected bid price 2999000, got %f", *(*board.Bids)[0].Price)
	}
	if *(*board.Bids)[0].Size != 0.1 {
		t.Errorf("Expected bid size 0.1, got %f", *(*board.Bids)[0].Size)
	}

	if len(*board.Asks) != 1 {
		t.Errorf("Expected 1 ask, got %d", len(*board.Asks))
	}
	if *(*board.Asks)[0].Price != 3001000 {
		t.Errorf("Expected ask price 3001000, got %f", *(*board.Asks)[0].Price)
	}
	if *(*board.Asks)[0].Size != 0.2 {
		t.Errorf("Expected ask size 0.2, got %f", *(*board.Asks)[0].Size)
	}
}

func TestGetExecutions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getexecutions" {
			t.Errorf("Expected path /v1/getexecutions, got %s", r.URL.Path)
		}

		executions := []MarketExecution{
			{
				Id:    i(1234),
				Side:  str("BUY"),
				Price: f32(3000000),
				Size:  f32(0.1),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(executions); err != nil {
			t.Fatalf("Failed to encode executions: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetexecutionsWithResponse(ctx, &GetV1GetexecutionsParams{
		ProductCode: "BTC_JPY",
		Count:       i(100),
		Before:      i(0),
		After:       i(0),
	})
	if err != nil {
		t.Fatalf("Failed to get executions: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected executions response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 execution, got %d", len(*resp.JSON200))
	}

	execution := (*resp.JSON200)[0]
	if *execution.Side != "BUY" {
		t.Errorf("Expected BUY side, got %s", *execution.Side)
	}
	if *execution.Price != 3000000 {
		t.Errorf("Expected price 3000000, got %f", *execution.Price)
	}
	if *execution.Size != 0.1 {
		t.Errorf("Expected size 0.1, got %f", *execution.Size)
	}
}

func TestPostSendParentOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/me/sendparentorder" {
			t.Errorf("Expected path /v1/me/sendparentorder, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"parent_order_acceptance_id": "JRF20150707-050237-639234"}`))
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	param := ParentOrderParameter{
		ProductCode:   "BTC_JPY",
		ConditionType: conditionType("LIMIT"),
		Side:          side("BUY"),
		Price:         f32(3000000),
		Size:          float32(0.1),
	}

	body := PostV1MeSendparentorderJSONRequestBody{
		OrderMethod: orderMethod("SIMPLE"),
		Parameters:  []ParentOrderParameter{param},
	}
	resp, err := client.Client().PostV1MeSendparentorderWithResponse(ctx, body)
	if err != nil {
		t.Fatalf("Failed to send parent order: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected parent order response, got nil")
	}

	if *resp.JSON200.ParentOrderAcceptanceId != "JRF20150707-050237-639234" {
		t.Errorf("Expected acceptance ID JRF20150707-050237-639234, got %s", *resp.JSON200.ParentOrderAcceptanceId)
	}
}

func TestPostCancelChildOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/me/cancelchildorder" {
			t.Errorf("Expected path /v1/me/cancelchildorder, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().PostV1MeCancelchildorderWithResponse(ctx, PostV1MeCancelchildorderJSONRequestBody{
		ProductCode:  "BTC_JPY",
		ChildOrderId: str("JRF20150707-050237-639234"),
	})
	if err != nil {
		t.Fatalf("Failed to cancel child order: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
	}
}

func TestGetTicker(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getticker" {
			t.Errorf("Expected path /v1/getticker, got %s", r.URL.Path)
		}

		ticker := Ticker{
			ProductCode:     str("BTC_JPY"),
			State:           str("RUNNING"),
			Timestamp:       func() *time.Time { t := time.Date(2025, 4, 4, 12, 0, 0, 0, time.UTC); return &t }(),
			BestBid:         f32(3000000),
			BestAsk:         f32(3000500),
			BestBidSize:     f32(0.1),
			BestAskSize:     f32(0.2),
			TotalBidDepth:   f32(100),
			TotalAskDepth:   f32(100),
			Ltp:             f32(3000000),
			Volume:          f32(50.0),
			VolumeByProduct: f32(50.0),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ticker); err != nil {
			t.Fatalf("Failed to encode ticker: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GettickerWithResponse(ctx, &GetV1GettickerParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get ticker: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected ticker response, got nil")
	}

	ticker := resp.JSON200
	if *ticker.ProductCode != "BTC_JPY" {
		t.Errorf("Expected product code BTC_JPY, got %s", *ticker.ProductCode)
	}
	if *ticker.State != "RUNNING" {
		t.Errorf("Expected state RUNNING, got %s", *ticker.State)
	}
	if *ticker.BestBid != 3000000 {
		t.Errorf("Expected best bid 3000000, got %f", *ticker.BestBid)
	}
	if *ticker.BestAsk != 3000500 {
		t.Errorf("Expected best ask 3000500, got %f", *ticker.BestAsk)
	}
	if *ticker.Volume != 50.0 {
		t.Errorf("Expected volume 50.0, got %f", *ticker.Volume)
	}
}

func TestGetBoardState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getboardstate" {
			t.Errorf("Expected path /v1/getboardstate, got %s", r.URL.Path)
		}

		state := BoardState{
			Health: boardStateHealth("NORMAL"),
			State:  boardStateState("RUNNING"),
			Data: &struct {
				SpecialQuotation *float32 `json:"special_quotation,omitempty"`
			}{
				SpecialQuotation: f32(3000000),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(state); err != nil {
			t.Fatalf("Failed to encode board state: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetboardstateWithResponse(ctx, &GetV1GetboardstateParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get board state: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected board state response, got nil")
	}

	state := resp.JSON200
	if *state.Health != "NORMAL" {
		t.Errorf("Expected health NORMAL, got %s", *state.Health)
	}
	if *state.State != "RUNNING" {
		t.Errorf("Expected state RUNNING, got %s", *state.State)
	}
	if *state.Data.SpecialQuotation != 3000000 {
		t.Errorf("Expected special quotation 3000000, got %f", *state.Data.SpecialQuotation)
	}
}

func TestGetHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/gethealth" {
			t.Errorf("Expected path /v1/gethealth, got %s", r.URL.Path)
		}

		health := ExchangeHealth{
			Status: (*ExchangeHealthStatus)(str("NORMAL")),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(health); err != nil {
			t.Fatalf("Failed to encode health: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GethealthWithResponse(ctx, &GetV1GethealthParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected health response, got nil")
	}

	health := resp.JSON200
	if *health.Status != "NORMAL" {
		t.Errorf("Expected status NORMAL, got %s", *health.Status)
	}
}

func TestGetFundingRate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getfundingrate" {
			t.Errorf("Expected path /v1/getfundingrate, got %s", r.URL.Path)
		}

		rate := FundingRate{
			CurrentFundingRate: f32(0.0001),
			NextFundingRateSettledate: func() *time.Time {
				t := time.Date(2025, 4, 4, 16, 0, 0, 0, time.UTC)
				return &t
			}(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(rate); err != nil {
			t.Fatalf("Failed to encode funding rate: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetfundingrateWithResponse(ctx, &GetV1GetfundingrateParams{
		ProductCode: "FX_BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get funding rate: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected funding rate response, got nil")
	}

	rate := resp.JSON200
	if *rate.CurrentFundingRate != 0.0001 {
		t.Errorf("Expected current funding rate 0.0001, got %f", *rate.CurrentFundingRate)
	}
}

func TestGetCorporateLeverage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getcorporateleverage" {
			t.Errorf("Expected path /v1/getcorporateleverage, got %s", r.URL.Path)
		}

		leverage := CorporateLeverage{
			CurrentMax: f32(4.0),
			CurrentStartdate: func() *time.Time {
				t := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			NextMax: f32(4.0),
			NextStartdate: func() *time.Time {
				t := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(leverage); err != nil {
			t.Fatalf("Failed to encode corporate leverage: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetcorporateleverageWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get corporate leverage: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected corporate leverage response, got nil")
	}

	leverage := resp.JSON200
	if *leverage.CurrentMax != 4.0 {
		t.Errorf("Expected current max leverage 4.0, got %f", *leverage.CurrentMax)
	}
	if *leverage.NextMax != 4.0 {
		t.Errorf("Expected next max leverage 4.0, got %f", *leverage.NextMax)
	}
}

func TestGetMarketsUSA(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getmarkets/usa" {
			t.Errorf("Expected path /v1/getmarkets/usa, got %s", r.URL.Path)
		}

		markets := []Market{
			{
				ProductCode: str("BTC_USD"),
				MarketType:  marketType("Spot"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(markets); err != nil {
			t.Fatalf("Failed to encode markets: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetmarketsUsaWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get USA markets: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected markets response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 market, got %d", len(*resp.JSON200))
	}

	market := (*resp.JSON200)[0]
	if *market.ProductCode != "BTC_USD" {
		t.Errorf("Expected BTC_USD, got %s", *market.ProductCode)
	}
	if *market.MarketType != "Spot" {
		t.Errorf("Expected Spot market type, got %s", *market.MarketType)
	}
}

func TestGetMarketsEU(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getmarkets/eu" {
			t.Errorf("Expected path /v1/getmarkets/eu, got %s", r.URL.Path)
		}

		markets := []Market{
			{
				ProductCode: str("BTC_EUR"),
				MarketType:  marketType("Spot"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(markets); err != nil {
			t.Fatalf("Failed to encode markets: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetmarketsEuWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get EU markets: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected markets response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 market, got %d", len(*resp.JSON200))
	}

	market := (*resp.JSON200)[0]
	if *market.ProductCode != "BTC_EUR" {
		t.Errorf("Expected BTC_EUR, got %s", *market.ProductCode)
	}
	if *market.MarketType != "Spot" {
		t.Errorf("Expected Spot market type, got %s", *market.MarketType)
	}
}

func TestGetChats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getchats" {
			t.Errorf("Expected path /v1/getchats, got %s", r.URL.Path)
		}

		messages := []ChatMessage{
			{
				Nickname: str("trader1"),
				Message:  str("Hello"),
				Date: func() *time.Time {
					t := time.Date(2025, 4, 4, 12, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			t.Fatalf("Failed to encode chat messages: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetchatsWithResponse(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get chats: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected chat messages response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 message, got %d", len(*resp.JSON200))
	}

	message := (*resp.JSON200)[0]
	if *message.Nickname != "trader1" {
		t.Errorf("Expected nickname trader1, got %s", *message.Nickname)
	}
	if *message.Message != "Hello" {
		t.Errorf("Expected message Hello, got %s", *message.Message)
	}
}

func TestGetChatsUSA(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getchats/usa" {
			t.Errorf("Expected path /v1/getchats/usa, got %s", r.URL.Path)
		}

		messages := []ChatMessage{
			{
				Nickname: str("trader_usa"),
				Message:  str("Hello from USA"),
				Date: func() *time.Time {
					t := time.Date(2025, 4, 4, 12, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			t.Fatalf("Failed to encode chat messages: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetchatsUsaWithResponse(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get USA chats: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected chat messages response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 message, got %d", len(*resp.JSON200))
	}

	message := (*resp.JSON200)[0]
	if *message.Nickname != "trader_usa" {
		t.Errorf("Expected nickname trader_usa, got %s", *message.Nickname)
	}
	if *message.Message != "Hello from USA" {
		t.Errorf("Expected message Hello from USA, got %s", *message.Message)
	}
}

func TestGetChatsEU(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getchats/eu" {
			t.Errorf("Expected path /v1/getchats/eu, got %s", r.URL.Path)
		}

		messages := []ChatMessage{
			{
				Nickname: str("trader_eu"),
				Message:  str("Hello from EU"),
				Date: func() *time.Time {
					t := time.Date(2025, 4, 4, 12, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			t.Fatalf("Failed to encode chat messages: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1GetchatsEuWithResponse(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get EU chats: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected chat messages response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 message, got %d", len(*resp.JSON200))
	}

	message := (*resp.JSON200)[0]
	if *message.Nickname != "trader_eu" {
		t.Errorf("Expected nickname trader_eu, got %s", *message.Nickname)
	}
	if *message.Message != "Hello from EU" {
		t.Errorf("Expected message Hello from EU, got %s", *message.Message)
	}
}

func TestMarkets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/markets" {
			t.Errorf("Expected path /v1/markets, got %s", r.URL.Path)
		}

		markets := []Market{
			{
				ProductCode: str("BTC_JPY"),
				MarketType:  marketType("Spot"),
			},
			{
				ProductCode: str("FX_BTC_JPY"),
				MarketType:  marketType("FX"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(markets); err != nil {
			t.Fatalf("Failed to encode markets: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1MarketsWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get markets: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected markets response, got nil")
	}

	if len(*resp.JSON200) != 2 {
		t.Errorf("Expected 2 markets, got %d", len(*resp.JSON200))
	}

	market := (*resp.JSON200)[0]
	if *market.ProductCode != "BTC_JPY" {
		t.Errorf("Expected BTC_JPY, got %s", *market.ProductCode)
	}
	if *market.MarketType != "Spot" {
		t.Errorf("Expected Spot market type, got %s", *market.MarketType)
	}
}

func TestMarketsUSA(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/markets/usa" {
			t.Errorf("Expected path /v1/markets/usa, got %s", r.URL.Path)
		}

		markets := []Market{
			{
				ProductCode: str("BTC_USD"),
				MarketType:  marketType("Spot"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(markets); err != nil {
			t.Fatalf("Failed to encode markets: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1MarketsUsaWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get USA markets: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected markets response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 market, got %d", len(*resp.JSON200))
	}

	market := (*resp.JSON200)[0]
	if *market.ProductCode != "BTC_USD" {
		t.Errorf("Expected BTC_USD, got %s", *market.ProductCode)
	}
	if *market.MarketType != "Spot" {
		t.Errorf("Expected Spot market type, got %s", *market.MarketType)
	}
}

func TestMarketsEU(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/markets/eu" {
			t.Errorf("Expected path /v1/markets/eu, got %s", r.URL.Path)
		}

		markets := []Market{
			{
				ProductCode: str("BTC_EUR"),
				MarketType:  marketType("Spot"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(markets); err != nil {
			t.Fatalf("Failed to encode markets: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1MarketsEuWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get EU markets: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected markets response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 market, got %d", len(*resp.JSON200))
	}

	market := (*resp.JSON200)[0]
	if *market.ProductCode != "BTC_EUR" {
		t.Errorf("Expected BTC_EUR, got %s", *market.ProductCode)
	}
	if *market.MarketType != "Spot" {
		t.Errorf("Expected Spot market type, got %s", *market.MarketType)
	}
}

func TestBoard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/board" {
			t.Errorf("Expected path /v1/board, got %s", r.URL.Path)
		}

		entries := []BoardEntry{
			{
				Price: f32(2999000),
				Size:  f32(0.1),
			},
		}

		askEntries := []BoardEntry{
			{
				Price: f32(3001000),
				Size:  f32(0.2),
			},
		}

		board := Board{
			MidPrice: f32(3000000),
			Bids:     &entries,
			Asks:     &askEntries,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(board); err != nil {
			t.Fatalf("Failed to encode board: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1BoardWithResponse(ctx, &GetV1BoardParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get board: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected board response, got nil")
	}

	board := resp.JSON200
	if *board.MidPrice != 3000000 {
		t.Errorf("Expected mid price 3000000, got %f", *board.MidPrice)
	}

	if len(*board.Bids) != 1 {
		t.Errorf("Expected 1 bid, got %d", len(*board.Bids))
	}
	if *(*board.Bids)[0].Price != 2999000 {
		t.Errorf("Expected bid price 2999000, got %f", *(*board.Bids)[0].Price)
	}
	if *(*board.Bids)[0].Size != 0.1 {
		t.Errorf("Expected bid size 0.1, got %f", *(*board.Bids)[0].Size)
	}

	if len(*board.Asks) != 1 {
		t.Errorf("Expected 1 ask, got %d", len(*board.Asks))
	}
	if *(*board.Asks)[0].Price != 3001000 {
		t.Errorf("Expected ask price 3001000, got %f", *(*board.Asks)[0].Price)
	}
	if *(*board.Asks)[0].Size != 0.2 {
		t.Errorf("Expected ask size 0.2, got %f", *(*board.Asks)[0].Size)
	}
}

func TestTicker(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ticker" {
			t.Errorf("Expected path /v1/ticker, got %s", r.URL.Path)
		}

		ticker := Ticker{
			ProductCode:     str("BTC_JPY"),
			State:           str("RUNNING"),
			Timestamp:       func() *time.Time { t := time.Date(2025, 4, 4, 12, 0, 0, 0, time.UTC); return &t }(),
			BestBid:         f32(3000000),
			BestAsk:         f32(3000500),
			BestBidSize:     f32(0.1),
			BestAskSize:     f32(0.2),
			TotalBidDepth:   f32(100),
			TotalAskDepth:   f32(100),
			Ltp:             f32(3000000),
			Volume:          f32(50.0),
			VolumeByProduct: f32(50.0),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ticker); err != nil {
			t.Fatalf("Failed to encode ticker: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1TickerWithResponse(ctx, &GetV1TickerParams{
		ProductCode: "BTC_JPY",
	})
	if err != nil {
		t.Fatalf("Failed to get ticker: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected ticker response, got nil")
	}

	ticker := resp.JSON200
	if *ticker.ProductCode != "BTC_JPY" {
		t.Errorf("Expected product code BTC_JPY, got %s", *ticker.ProductCode)
	}
	if *ticker.State != "RUNNING" {
		t.Errorf("Expected state RUNNING, got %s", *ticker.State)
	}
	if *ticker.BestBid != 3000000 {
		t.Errorf("Expected best bid 3000000, got %f", *ticker.BestBid)
	}
	if *ticker.BestAsk != 3000500 {
		t.Errorf("Expected best ask 3000500, got %f", *ticker.BestAsk)
	}
	if *ticker.Volume != 50.0 {
		t.Errorf("Expected volume 50.0, got %f", *ticker.Volume)
	}
}

func TestExecutions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/executions" {
			t.Errorf("Expected path /v1/executions, got %s", r.URL.Path)
		}

		executions := []MarketExecution{
			{
				Id:    i(1234),
				Side:  str("BUY"),
				Price: f32(3000000),
				Size:  f32(0.1),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(executions); err != nil {
			t.Fatalf("Failed to encode executions: %v", err)
		}
	}))
	defer srv.Close()

	client, err := NewAuthenticatedClient(auth.APICredentials{}, srv.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Client().GetV1ExecutionsWithResponse(ctx, &GetV1ExecutionsParams{
		ProductCode: "BTC_JPY",
		Count:       i(100),
		Before:      i(0),
		After:       i(0),
	})
	if err != nil {
		t.Fatalf("Failed to get executions: %v", err)
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected executions response, got nil")
	}

	if len(*resp.JSON200) != 1 {
		t.Errorf("Expected 1 execution, got %d", len(*resp.JSON200))
	}

	execution := (*resp.JSON200)[0]
	if *execution.Side != "BUY" {
		t.Errorf("Expected BUY side, got %s", *execution.Side)
	}
	if *execution.Price != 3000000 {
		t.Errorf("Expected price 3000000, got %f", *execution.Price)
	}
	if *execution.Size != 0.1 {
		t.Errorf("Expected size 0.1, got %f", *execution.Size)
	}
}

// Error case tests
func TestPrivateAPIErrorCases(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{"InvalidBalance", "/v1/me/getbalance"},
		{"InvalidParentOrders", "/v1/me/getparentorders"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Errorf("Expected path %s, got %s", tc.path, r.URL.Path)
				}
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte(`{"error": "invalid authentication"}`))
				if err != nil {
					t.Fatalf("Failed to write error response: %v", err)
				}
			}))
			defer srv.Close()

			client, err := NewAuthenticatedClient(auth.APICredentials{
				APIKey:    "invalid-key",
				APISecret: "invalid-secret",
			}, srv.URL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			ctx := context.Background()
			var statusCode int

			switch tc.path {
			case "/v1/me/getbalance":
				resp, _ := client.Client().GetV1MeGetbalanceWithResponse(ctx)
				statusCode = resp.StatusCode()
			case "/v1/me/getparentorders":
				resp, _ := client.Client().GetV1MeGetparentordersWithResponse(ctx, &GetV1MeGetparentordersParams{})
				statusCode = resp.StatusCode()
			}

			if statusCode != http.StatusUnauthorized {
				t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, statusCode)
			}
		})
	}
}
