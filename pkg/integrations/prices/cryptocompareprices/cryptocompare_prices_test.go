package cryptocompareprices

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
			resp := map[string]float64{"USD": 87000.0}
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
		assert.Equal(t, 87000.0, price.Value)
	}
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]map[string]float64{
				"BTC": {"USD": 87000.0},
				"ETH": {"USD": 2900.0},
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

	price := &prices.Price{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestPriceFetcher_FetchAll(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"Data": []map[string]interface{}{
					{
						"CoinInfo": map[string]string{
							"Name":     "BTC",
							"FullName": "Bitcoin",
						},
						"RAW": map[string]interface{}{
							"USD": map[string]float64{
								"PRICE": 87000.0,
							},
						},
					},
					{
						"CoinInfo": map[string]string{
							"Name":     "ETH",
							"FullName": "Ethereum",
						},
						"RAW": map[string]interface{}{
							"USD": map[string]float64{
								"PRICE": 2900.0,
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
			t.Logf("%s (%s): %f", p.Asset.Name, p.Asset.Symbol, p.Value)
		}
	} else {
		assert.Len(t, allPrices, 2)
		assert.Equal(t, "BTC", allPrices[0].Asset.Symbol)
		assert.Equal(t, "Bitcoin", allPrices[0].Asset.Name)
		assert.Equal(t, 87000.0, allPrices[0].Value)
	}
}

func TestPriceFetcher_WithAPIKey(t *testing.T) {
	if isIntegration() {
		t.Skip("skipping API key test in integration mode")
	}

	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("authorization")
		resp := map[string]float64{"USD": 87000.0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcherWithKey("test-api-key")
	fetcher.BaseURL = server.URL

	price := &prices.Price{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)
	require.NoError(t, err)

	assert.Equal(t, "Apikey test-api-key", receivedAuth)
}
