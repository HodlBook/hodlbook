package defillamaprices

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
				"coins": map[string]interface{}{
					"coingecko:bitcoin": map[string]interface{}{
						"price":     87000.0,
						"symbol":    "BTC",
						"timestamp": 1234567890,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	price := &prices.Price{Asset: prices.Asset{Name: "bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)
	require.NoError(t, err)

	if isIntegration() {
		assert.Greater(t, price.Value, 0.0)
		t.Logf("BTC/USD: %f", price.Value)
	} else {
		assert.Equal(t, 87000.0, price.Value)
	}
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"coins": map[string]interface{}{
					"coingecko:bitcoin": map[string]interface{}{
						"price":  87000.0,
						"symbol": "BTC",
					},
					"coingecko:ethereum": map[string]interface{}{
						"price":  2900.0,
						"symbol": "ETH",
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
		{Asset: prices.Asset{Name: "bitcoin", Symbol: "BTC"}},
		{Asset: prices.Asset{Name: "ethereum", Symbol: "ETH"}},
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

	price := &prices.Price{Asset: prices.Asset{Name: "bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestPriceFetcher_FetchAll_NotSupported(t *testing.T) {
	fetcher := NewPriceFetcher()
	_, err := fetcher.FetchAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}
