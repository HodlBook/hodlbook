package krakenprices

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
			resp := tickerResponse{
				Error: []string{},
				Result: map[string]tickerResultEntry{
					"XXBTZUSD": {
						Close: []string{"87267.53", "0.001"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	price := &prices.Price{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)
	require.NoError(t, err)

	if isIntegration() {
		assert.Greater(t, price.Value, 0.0)
		t.Logf("BTC/USD: %f", price.Value)
	} else {
		assert.Equal(t, 87267.53, price.Value)
	}
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := tickerResponse{
				Error: []string{},
				Result: map[string]tickerResultEntry{
					"XXBTZUSD": {Close: []string{"87222.51", "0.001"}},
					"XETHZUSD": {Close: []string{"2933.91", "0.01"}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	testPrices := []*prices.Price{
		{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}},
		{Asset: prices.Asset{Name: "Ethereum", Symbol: "ETH"}},
	}

	err := fetcher.FetchMany(testPrices...)
	require.NoError(t, err)

	if isIntegration() {
		assert.Greater(t, testPrices[0].Value, 0.0)
		assert.Greater(t, testPrices[1].Value, 0.0)
		for _, p := range testPrices {
			t.Logf("%s/USD: %f", p.Asset.Symbol, p.Value)
		}
	} else {
		assert.Equal(t, 87222.51, testPrices[0].Value)
		assert.Equal(t, 2933.91, testPrices[1].Value)
	}
}

func TestPriceFetcher_FetchMany_StablecoinsReturnOne(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := tickerResponse{
				Error:  []string{},
				Result: map[string]tickerResultEntry{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

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

	price := &prices.Price{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestPriceFetcher_Fetch_APIError(t *testing.T) {
	if isIntegration() {
		t.Skip("skipping API error test in integration mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := tickerResponse{
			Error:  []string{"EQuery:Unknown asset pair"},
			Result: map[string]tickerResultEntry{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	price := &prices.Price{Asset: prices.Asset{Name: "Unknown", Symbol: "XXX"}}
	err := fetcher.Fetch(price)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown asset pair")
}

func TestPriceFetcher_FetchAll(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := tickerResponse{
				Error: []string{},
				Result: map[string]tickerResultEntry{
					"XXBTZUSD": {Close: []string{"87000.0", "0.001"}},
					"XETHZUSD": {Close: []string{"2900.0", "0.01"}},
					"XXBTZEUR": {Close: []string{"80000.0", "0.001"}},
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
		priceMap := make(map[string]float64)
		for _, p := range allPrices {
			priceMap[p.Asset.Symbol] = p.Value
		}
		assert.Equal(t, 87000.0, priceMap["BTC"])
		assert.Equal(t, 2900.0, priceMap["ETH"])
	}
}

func TestToKrakenPair(t *testing.T) {
	assert.Equal(t, "XXBTZUSD", toKrakenPair("BTC"))
	assert.Equal(t, "XETHZUSD", toKrakenPair("ETH"))
	assert.Equal(t, "SOLUSD", toKrakenPair("SOL"))
}

func TestFromKrakenPair(t *testing.T) {
	assert.Equal(t, "BTC", fromKrakenPair("XXBTZUSD"))
	assert.Equal(t, "ETH", fromKrakenPair("XETHZUSD"))
	assert.Equal(t, "SOL", fromKrakenPair("SOLUSD"))
}
