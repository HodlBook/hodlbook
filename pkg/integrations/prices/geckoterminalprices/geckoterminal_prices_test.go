package geckoterminalprices

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"hodlbook/pkg/types/prices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isIntegration() bool {
	return os.Getenv("INTEGRATION") == "true"
}

func TestPriceFetcher_Fetch(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"attributes": map[string]interface{}{
							"base_token_price_usd": "87000.50",
							"name":                 "WETH / USDC",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	price := &prices.Price{Asset: prices.Asset{Name: "Wrapped Ether", Symbol: "WETH"}}
	err := fetcher.Fetch(price)
	require.NoError(t, err)

	if isIntegration() {
		assert.Greater(t, price.Value, 0.0)
		t.Logf("WETH/USD: %f", price.Value)
	} else {
		assert.Equal(t, 87000.50, price.Value)
	}
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			var priceStr string
			if callCount == 1 {
				priceStr = "87000.0"
			} else {
				priceStr = "2900.0"
			}
			resp := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"attributes": map[string]interface{}{
							"base_token_price_usd": priceStr,
							"name":                 "TOKEN / USDC",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	testPrices := []*prices.Price{
		{Asset: prices.Asset{Name: "Wrapped Ether", Symbol: "WETH"}},
		{Asset: prices.Asset{Name: "Wrapped Bitcoin", Symbol: "WBTC"}},
	}

	err := fetcher.FetchMany(testPrices...)
	require.NoError(t, err)

	if isIntegration() {
		for _, p := range testPrices {
			if p.Value > 0 {
				t.Logf("%s/USD: %f", p.Asset.Symbol, p.Value)
			}
		}
	} else {
		assert.Equal(t, 87000.0, testPrices[0].Value)
		assert.Equal(t, 2900.0, testPrices[1].Value)
	}
}

func TestPriceFetcher_FetchMany_StablecoinsReturnOne(t *testing.T) {
	fetcher := NewPriceFetcher()

	testPrices := []*prices.Price{
		{Asset: prices.Asset{Name: "USD", Symbol: "USD"}},
		{Asset: prices.Asset{Name: "Tether", Symbol: "USDT"}},
		{Asset: prices.Asset{Name: "USD Coin", Symbol: "USDC"}},
	}

	err := fetcher.FetchMany(testPrices...)
	require.NoError(t, err)

	for _, pair := range testPrices {
		assert.Equal(t, 1.0, pair.Value, "stablecoin %s should be 1.0", pair.Asset.Symbol)
	}
}

func TestPriceFetcher_Fetch_HTTPError(t *testing.T) {
	if isIntegration() {
		t.Skip("skipping HTTP error test in integration mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	price := &prices.Price{Asset: prices.Asset{Name: "Wrapped Ether", Symbol: "WETH"}}
	err := fetcher.Fetch(price)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestPriceFetcher_Fetch_NoPoolsFound(t *testing.T) {
	if isIntegration() {
		t.Skip("skipping no pools test in integration mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	price := &prices.Price{Asset: prices.Asset{Name: "Unknown", Symbol: "UNKNOWN"}}
	err := fetcher.Fetch(price)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pools found")
}

func TestPriceFetcher_FetchAll(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"attributes": map[string]interface{}{
							"name":                 "WETH / USDC",
							"base_token_price_usd": "2900.0",
						},
						"relationships": map[string]interface{}{
							"base_token": map[string]interface{}{
								"data": map[string]string{
									"id": "eth_0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
								},
							},
						},
					},
					{
						"attributes": map[string]interface{}{
							"name":                 "PEPE / WETH",
							"base_token_price_usd": "0.00001234",
						},
						"relationships": map[string]interface{}{
							"base_token": map[string]interface{}{
								"data": map[string]string{
									"id": "eth_0x6982508145454ce325ddbe47a25d4ec3d2311933",
								},
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	allPrices, err := fetcher.FetchAll()
	require.NoError(t, err)

	if isIntegration() {
		assert.Greater(t, len(allPrices), 0)
		t.Logf("fetched %d prices", len(allPrices))
		for _, p := range allPrices[:min(5, len(allPrices))] {
			t.Logf("%s: %f", p.Asset.Symbol, p.Value)
		}
	} else {
		assert.Len(t, allPrices, 2)
		assert.Equal(t, "WETH", allPrices[0].Asset.Symbol)
		assert.Equal(t, 2900.0, allPrices[0].Value)
	}
}

func TestNewPriceFetcherForNetwork(t *testing.T) {
	fetcher := NewPriceFetcherForNetwork("base")
	assert.Equal(t, "base", fetcher.Network)
}

func TestExtractSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"WETH / USDC", "WETH"},
		{"PEPE / WETH", "PEPE"},
		{"TOKEN", "TOKEN"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractSymbol(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
