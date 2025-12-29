package binanceprices

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
		resp := struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}{
			Symbol: "BTCUSD",
			Price:  "87267.53",
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
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}{
			{Symbol: "BTCUSDT", Price: "87222.51"},
			{Symbol: "ETHUSDT", Price: "2933.91"},
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
}

func TestPriceFetcher_FetchMany_StablecoinsReturnOne(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

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
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}{
			{Symbol: "BTCUSD", Price: "87000.0"},
			{Symbol: "ETHUSDT", Price: "2900.0"},
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
	assert.Equal(t, "BTC", allPrices[0].Asset.Symbol)
	assert.Equal(t, 87000.0, allPrices[0].Value)
	assert.Equal(t, "ETH", allPrices[1].Asset.Symbol)
	assert.Equal(t, 2900.0, allPrices[1].Value)
}
