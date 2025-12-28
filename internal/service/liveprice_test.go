package service

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"hodlbook/internal/models"
	"hodlbook/pkg/integrations/memcache"
	pricesPkg "hodlbook/pkg/integrations/prices"
	"hodlbook/pkg/integrations/wmPubsub"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type mockAssetRepo struct {
	assets          []models.Asset
	exchangeSymbols []string
}

func (m *mockAssetRepo) GetAllAssets() ([]models.Asset, error) {
	return m.assets, nil
}

func (m *mockAssetRepo) GetUniqueExchangeSymbols() ([]string, error) {
	return m.exchangeSymbols, nil
}

func TestLivePriceService_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	cache := memcache.New[string, float64]()
	fetcher := pricesPkg.NewPriceService()
	ch := make(chan []byte, 10)
	pub := wmPubsub.New(wmPubsub.WithChannel(ch), wmPubsub.WithContext(ctx))
	repo := &mockAssetRepo{}

	tests := []struct {
		name string
		opts []LivePriceOption
	}{
		{"no context", []LivePriceOption{
			WithLivePriceLogger(discardLogger),
			WithLivePriceCache(cache),
			WithLivePriceFetcher(fetcher),
			WithLivePricePublisher(pub),
			WithLivePriceRepo(repo),
		}},
		{"no logger", []LivePriceOption{
			WithLivePriceContext(ctx),
			WithLivePriceCache(cache),
			WithLivePriceFetcher(fetcher),
			WithLivePricePublisher(pub),
			WithLivePriceRepo(repo),
		}},
		{"no cache", []LivePriceOption{
			WithLivePriceContext(ctx),
			WithLivePriceLogger(discardLogger),
			WithLivePriceFetcher(fetcher),
			WithLivePricePublisher(pub),
			WithLivePriceRepo(repo),
		}},
		{"no fetcher", []LivePriceOption{
			WithLivePriceContext(ctx),
			WithLivePriceLogger(discardLogger),
			WithLivePriceCache(cache),
			WithLivePricePublisher(pub),
			WithLivePriceRepo(repo),
		}},
		{"no publisher", []LivePriceOption{
			WithLivePriceContext(ctx),
			WithLivePriceLogger(discardLogger),
			WithLivePriceCache(cache),
			WithLivePriceFetcher(fetcher),
			WithLivePriceRepo(repo),
		}},
		{"no repo", []LivePriceOption{
			WithLivePriceContext(ctx),
			WithLivePriceLogger(discardLogger),
			WithLivePriceCache(cache),
			WithLivePriceFetcher(fetcher),
			WithLivePricePublisher(pub),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLivePriceService(tt.opts...)
			assert.Error(t, err)
		})
	}
}

func TestLivePriceService_SyncFromDB(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := memcache.New[string, float64]()
	fetcher := pricesPkg.NewPriceService()
	ch := make(chan []byte, 10)
	pub := wmPubsub.New(wmPubsub.WithChannel(ch), wmPubsub.WithContext(ctx))
	repo := &mockAssetRepo{
		assets: []models.Asset{
			{ID: 1, Symbol: "BTC", Name: "Bitcoin"},
			{ID: 2, Symbol: "ETH", Name: "Ethereum"},
		},
	}

	svc, err := NewLivePriceService(
		WithLivePriceContext(ctx),
		WithLivePriceLogger(discardLogger),
		WithLivePriceCache(cache),
		WithLivePriceFetcher(fetcher),
		WithLivePricePublisher(pub),
		WithLivePriceRepo(repo),
	)
	require.NoError(t, err)

	err = svc.syncFromDB()
	require.NoError(t, err)

	assert.Equal(t, 2, cache.Len())
	_, ok := cache.Get("BTC")
	assert.True(t, ok)
	_, ok = cache.Get("ETH")
	assert.True(t, ok)
}

func TestLivePriceService_FetchAndPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := memcache.New[string, float64]()
	cache.Set("BTC", 0)
	cache.Set("ETH", 0)

	fetcher := pricesPkg.NewPriceService()
	ch := make(chan []byte, 10)
	pub := wmPubsub.New(wmPubsub.WithChannel(ch), wmPubsub.WithContext(ctx))
	repo := &mockAssetRepo{}

	svc, err := NewLivePriceService(
		WithLivePriceContext(ctx),
		WithLivePriceLogger(discardLogger),
		WithLivePriceCache(cache),
		WithLivePriceFetcher(fetcher),
		WithLivePricePublisher(pub),
		WithLivePriceRepo(repo),
	)
	require.NoError(t, err)

	err = svc.fetchAndPublish()
	require.NoError(t, err)

	btcPrice, ok := cache.Get("BTC")
	assert.True(t, ok)
	assert.Greater(t, btcPrice, float64(0))

	ethPrice, ok := cache.Get("ETH")
	assert.True(t, ok)
	assert.Greater(t, ethPrice, float64(0))

	select {
	case data := <-ch:
		var prices map[string]float64
		err := json.Unmarshal(data, &prices)
		require.NoError(t, err)
		assert.Contains(t, prices, "BTC")
		assert.Contains(t, prices, "ETH")
		assert.Greater(t, prices["BTC"], float64(0))
		assert.Greater(t, prices["ETH"], float64(0))
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive published prices")
	}
}

func TestLivePriceService_CacheAccess(t *testing.T) {
	cache := memcache.New[string, float64]()
	cache.Set("BTC", 50000.0)
	cache.Set("ETH", 3000.0)

	price, ok := cache.Get("BTC")
	assert.True(t, ok)
	assert.Equal(t, 50000.0, price)

	_, ok = cache.Get("UNKNOWN")
	assert.False(t, ok)

	keys := cache.Keys()
	assert.Len(t, keys, 2)
}
