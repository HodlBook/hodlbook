package coingeckoprices

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hodlbook/pkg/types/prices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceFetcher_Fetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]map[string]float64{
			"bitcoin": {"usd": 87267.53},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	price := &prices.Price{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)
	require.NoError(t, err)

	assert.Equal(t, 87267.53, price.Value)
	t.Logf("%sUSD: %f", price.Asset.Symbol, price.Value)
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]map[string]float64{
			"bitcoin":  {"usd": 87222.51},
			"ethereum": {"usd": 2933.91},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	testPrices := []*prices.Price{
		{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}},
		{Asset: prices.Asset{Name: "Ethereum", Symbol: "ETH"}},
	}

	err := fetcher.FetchMany(testPrices...)
	require.NoError(t, err)

	assert.Equal(t, 87222.51, testPrices[0].Value)
	assert.Equal(t, 2933.91, testPrices[1].Value)

	for _, pair := range testPrices {
		t.Logf("%sUSD: %f", pair.Asset.Symbol, pair.Value)
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

func TestPriceFetcher_FetchAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []struct {
			ID     string  `json:"id"`
			Symbol string  `json:"symbol"`
			Price  float64 `json:"current_price"`
		}{
			{ID: "bitcoin", Symbol: "btc", Price: 87000.0},
			{ID: "ethereum", Symbol: "eth", Price: 2900.0},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	allPrices, err := fetcher.FetchAll()
	require.NoError(t, err)

	assert.Len(t, allPrices, 2)
	assert.Equal(t, "bitcoin", allPrices[0].Asset.Name)
	assert.Equal(t, "BTC", allPrices[0].Asset.Symbol)
	assert.Equal(t, 87000.0, allPrices[0].Value)
}
