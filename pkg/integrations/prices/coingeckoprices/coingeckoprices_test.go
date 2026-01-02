package coingeckoprices

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"hodlbook/pkg/types/prices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	isIntegration = os.Getenv("INTEGRATION") == "true"
)

func TestPriceFetcher_Fetch(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]map[string]float64{
				"bitcoin": {"usd": 87267.53},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	price := &prices.Price{Asset: prices.Asset{Name: "Bitcoin", Symbol: "BTC"}}
	err := fetcher.Fetch(price)
	if isIntegration && err != nil && strings.Contains(err.Error(), "429") {
		t.Skip("skipping due to rate limiting (429)")
	}
	require.NoError(t, err)

	if isIntegration {
		assert.Greater(t, price.Value, 0.0)
		t.Logf("BTC/USD: %f", price.Value)
	} else {
		assert.Equal(t, 87267.53, price.Value)
	}
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if !isIntegration {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]map[string]float64{
				"bitcoin":  {"usd": 87222.51},
				"ethereum": {"usd": 2933.91},
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
	if isIntegration && err != nil && strings.Contains(err.Error(), "429") {
		t.Skip("skipping due to rate limiting (429)")
	}
	require.NoError(t, err)

	if isIntegration {
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
	if isIntegration {
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

	if !isIntegration {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			var resp []struct {
				ID     string  `json:"id"`
				Symbol string  `json:"symbol"`
				Price  float64 `json:"current_price"`
			}
			switch page {
			case "1", "":
				resp = []struct {
					ID     string  `json:"id"`
					Symbol string  `json:"symbol"`
					Price  float64 `json:"current_price"`
				}{
					{ID: "bitcoin", Symbol: "btc", Price: 87000.0},
					{ID: "ethereum", Symbol: "eth", Price: 2900.0},
				}
			case "2":
				resp = []struct {
					ID     string  `json:"id"`
					Symbol string  `json:"symbol"`
					Price  float64 `json:"current_price"`
				}{
					{ID: "lombard-protocol", Symbol: "bard", Price: 0.78},
					{ID: "spectra-finance", Symbol: "spectra", Price: 0.006},
				}
			default:
				resp = []struct {
					ID     string  `json:"id"`
					Symbol string  `json:"symbol"`
					Price  float64 `json:"current_price"`
				}{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher.BaseURL = server.URL
	}

	allPrices, err := fetcher.FetchAll()
	if isIntegration && err != nil && strings.Contains(err.Error(), "429") {
		t.Skip("skipping due to rate limiting (429)")
	}
	require.NoError(t, err)

	if isIntegration {
		// defaultPages=5, 250/page = up to 1250 assets
		assert.Greater(t, len(allPrices), 500)
		assert.LessOrEqual(t, len(allPrices), 1250)
		t.Logf("fetched %d prices (5 pages)", len(allPrices))
		for _, p := range allPrices[:min(5, len(allPrices))] {
			t.Logf("%s (%s): %f", p.Asset.Name, p.Asset.Symbol, p.Value)
		}
	} else {
		assert.Len(t, allPrices, 4)
		assert.Equal(t, "bitcoin", allPrices[0].Asset.Name)
		assert.Equal(t, "BTC", allPrices[0].Asset.Symbol)
		assert.Equal(t, 87000.0, allPrices[0].Value)
		assert.Equal(t, "lombard-protocol", allPrices[2].Asset.Name)
		assert.Equal(t, "BARD", allPrices[2].Asset.Symbol)
	}
}

func TestPriceFetcher_FetchAllPages_StopsOnEmptyPage(t *testing.T) {
	pageRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageRequests++
		page := r.URL.Query().Get("page")
		var resp []struct {
			ID     string  `json:"id"`
			Symbol string  `json:"symbol"`
			Price  float64 `json:"current_price"`
		}
		if page == "1" || page == "" {
			resp = []struct {
				ID     string  `json:"id"`
				Symbol string  `json:"symbol"`
				Price  float64 `json:"current_price"`
			}{
				{ID: "bitcoin", Symbol: "btc", Price: 87000.0},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	allPrices, err := fetcher.FetchAllPages(5)
	require.NoError(t, err)
	assert.Len(t, allPrices, 1)
	assert.Equal(t, 2, pageRequests)
}

func TestPriceFetcher_FetchAllPages_FirstPageError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	_, err := fetcher.FetchAllPages(3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestPriceFetcher_FetchAllPages_LaterPageError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "1" || page == "" {
			resp := []struct {
				ID     string  `json:"id"`
				Symbol string  `json:"symbol"`
				Price  float64 `json:"current_price"`
			}{
				{ID: "bitcoin", Symbol: "btc", Price: 87000.0},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	allPrices, err := fetcher.FetchAllPages(3)
	require.NoError(t, err)
	assert.Len(t, allPrices, 1)
}

func TestPriceFetcher_FetchAllPages_SkipsNullPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1", "":
			w.Write([]byte(`[
				{"id": "bitcoin", "symbol": "btc", "current_price": 87000.0},
				{"id": "no-price-coin", "symbol": "npc", "current_price": null},
				{"id": "zero-price", "symbol": "zero", "current_price": 0},
				{"id": "ethereum", "symbol": "eth", "current_price": 2900.0}
			]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	allPrices, err := fetcher.FetchAllPages(2)
	require.NoError(t, err)
	assert.Len(t, allPrices, 2)
	assert.Equal(t, "bitcoin", allPrices[0].Asset.Name)
	assert.Equal(t, "ethereum", allPrices[1].Asset.Name)
}

func TestPriceFetcher_FetchAllPages_StopsOnAllNullPage(t *testing.T) {
	pageRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageRequests++
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1", "":
			w.Write([]byte(`[
				{"id": "bitcoin", "symbol": "btc", "current_price": 87000.0}
			]`))
		case "2":
			w.Write([]byte(`[
				{"id": "dead-coin-1", "symbol": "dc1", "current_price": null},
				{"id": "dead-coin-2", "symbol": "dc2", "current_price": 0}
			]`))
		case "3":
			w.Write([]byte(`[
				{"id": "shouldnt-reach", "symbol": "sr", "current_price": 100.0}
			]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer server.Close()

	fetcher := NewPriceFetcher()
	fetcher.BaseURL = server.URL

	allPrices, err := fetcher.FetchAllPages(5)
	require.NoError(t, err)
	assert.Len(t, allPrices, 1)
	assert.Equal(t, "bitcoin", allPrices[0].Asset.Name)
	assert.Equal(t, 2, pageRequests)
}
