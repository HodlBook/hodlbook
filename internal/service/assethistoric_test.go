package service

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"hodlbook/internal/models"
	pricesPkg "hodlbook/pkg/integrations/prices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var assetHistoricDiscardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type mockAssetHistoricRepo struct {
	values    map[string][]models.AssetHistoricValue
	mu        sync.Mutex
	insertErr error
}

func newMockAssetHistoricRepo() *mockAssetHistoricRepo {
	return &mockAssetHistoricRepo{
		values: make(map[string][]models.AssetHistoricValue),
	}
}

func (m *mockAssetHistoricRepo) SelectAllBySymbol(symbol string) ([]models.AssetHistoricValue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.values[symbol], nil
}

func (m *mockAssetHistoricRepo) Insert(value *models.AssetHistoricValue) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[value.Symbol] = append(m.values[value.Symbol], *value)
	return nil
}

func (m *mockAssetHistoricRepo) GetValues(symbol string) []models.AssetHistoricValue {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.values[symbol]
}

func TestAssetHistoricService_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	ch := make(chan []byte, 10)

	tests := []struct {
		name string
		opts []AssetHistoricOption
	}{
		{"no context", []AssetHistoricOption{
			WithAssetHistoricLogger(assetHistoricDiscardLogger),
			WithAssetHistoricFetcher(fetcher),
			WithAssetHistoricRepo(repo),
			WithAssetHistoricChannel(ch),
		}},
		{"no logger", []AssetHistoricOption{
			WithAssetHistoricContext(ctx),
			WithAssetHistoricFetcher(fetcher),
			WithAssetHistoricRepo(repo),
			WithAssetHistoricChannel(ch),
		}},
		{"no fetcher", []AssetHistoricOption{
			WithAssetHistoricContext(ctx),
			WithAssetHistoricLogger(assetHistoricDiscardLogger),
			WithAssetHistoricRepo(repo),
			WithAssetHistoricChannel(ch),
		}},
		{"no repo", []AssetHistoricOption{
			WithAssetHistoricContext(ctx),
			WithAssetHistoricLogger(assetHistoricDiscardLogger),
			WithAssetHistoricFetcher(fetcher),
			WithAssetHistoricChannel(ch),
		}},
		{"no channel", []AssetHistoricOption{
			WithAssetHistoricContext(ctx),
			WithAssetHistoricLogger(assetHistoricDiscardLogger),
			WithAssetHistoricFetcher(fetcher),
			WithAssetHistoricRepo(repo),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAssetHistoricService(tt.opts...)
			assert.Error(t, err)
		})
	}
}

func TestAssetHistoricService_ValidConfig(t *testing.T) {
	ctx := context.Background()
	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	ch := make(chan []byte, 10)

	svc, err := NewAssetHistoricService(
		WithAssetHistoricContext(ctx),
		WithAssetHistoricLogger(assetHistoricDiscardLogger),
		WithAssetHistoricFetcher(fetcher),
		WithAssetHistoricRepo(repo),
		WithAssetHistoricChannel(ch),
	)

	require.NoError(t, err)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.Publisher())
}

func TestAssetHistoricService_HandleAssetCreated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	ch := make(chan []byte, 10)

	svc, err := NewAssetHistoricService(
		WithAssetHistoricContext(ctx),
		WithAssetHistoricLogger(assetHistoricDiscardLogger),
		WithAssetHistoricFetcher(fetcher),
		WithAssetHistoricRepo(repo),
		WithAssetHistoricChannel(ch),
	)
	require.NoError(t, err)

	asset := models.Asset{
		ID:     1,
		Symbol: "BTC",
		Name:   "Bitcoin",
	}
	data, err := json.Marshal(asset)
	require.NoError(t, err)

	err = svc.handleAssetCreated(data)
	require.NoError(t, err)

	values := repo.GetValues("BTC")
	assert.Len(t, values, 1)
	assert.Equal(t, "BTC", values[0].Symbol)
	assert.Greater(t, values[0].Value, float64(0))
	assert.False(t, values[0].Timestamp.IsZero())
}

func TestAssetHistoricService_HandleAssetCreatedSkipsExisting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	repo.values["BTC"] = []models.AssetHistoricValue{
		{Symbol: "BTC", Value: 50000, Timestamp: time.Now()},
	}
	ch := make(chan []byte, 10)

	svc, err := NewAssetHistoricService(
		WithAssetHistoricContext(ctx),
		WithAssetHistoricLogger(assetHistoricDiscardLogger),
		WithAssetHistoricFetcher(fetcher),
		WithAssetHistoricRepo(repo),
		WithAssetHistoricChannel(ch),
	)
	require.NoError(t, err)

	asset := models.Asset{
		ID:     1,
		Symbol: "BTC",
		Name:   "Bitcoin",
	}
	data, err := json.Marshal(asset)
	require.NoError(t, err)

	err = svc.handleAssetCreated(data)
	require.NoError(t, err)

	values := repo.GetValues("BTC")
	assert.Len(t, values, 1)
}

func TestAssetHistoricService_HandleAssetCreatedInvalidJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	ch := make(chan []byte, 10)

	svc, err := NewAssetHistoricService(
		WithAssetHistoricContext(ctx),
		WithAssetHistoricLogger(assetHistoricDiscardLogger),
		WithAssetHistoricFetcher(fetcher),
		WithAssetHistoricRepo(repo),
		WithAssetHistoricChannel(ch),
	)
	require.NoError(t, err)

	err = svc.handleAssetCreated([]byte("invalid json"))
	assert.Error(t, err)
}

func TestAssetHistoricService_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	ch := make(chan []byte, 10)

	svc, err := NewAssetHistoricService(
		WithAssetHistoricContext(ctx),
		WithAssetHistoricLogger(assetHistoricDiscardLogger),
		WithAssetHistoricFetcher(fetcher),
		WithAssetHistoricRepo(repo),
		WithAssetHistoricChannel(ch),
	)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)
}

func TestAssetHistoricService_PublishAndHandle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := newMockAssetHistoricRepo()
	ch := make(chan []byte, 10)

	svc, err := NewAssetHistoricService(
		WithAssetHistoricContext(ctx),
		WithAssetHistoricLogger(assetHistoricDiscardLogger),
		WithAssetHistoricFetcher(fetcher),
		WithAssetHistoricRepo(repo),
		WithAssetHistoricChannel(ch),
	)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)

	asset := models.Asset{
		ID:     2,
		Symbol: "ETH",
		Name:   "Ethereum",
	}
	data, err := json.Marshal(asset)
	require.NoError(t, err)

	err = svc.Publisher().Publish(data)
	require.NoError(t, err)

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for asset historic value to be inserted")
		case <-ticker.C:
			values := repo.GetValues("ETH")
			if len(values) > 0 {
				assert.Equal(t, "ETH", values[0].Symbol)
				assert.Greater(t, values[0].Value, float64(0))
				return
			}
		}
	}
}
