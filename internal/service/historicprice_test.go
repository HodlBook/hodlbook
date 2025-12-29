package service

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"hodlbook/internal/models"
	pricesPkg "hodlbook/pkg/integrations/prices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var historicDiscardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type mockHistoricRepo struct {
	symbols         []string
	historicSymbols []string
	values          []models.AssetHistoricValue
	mu              sync.Mutex
	insertErr       error
}

func (m *mockHistoricRepo) GetUniqueSymbols() ([]string, error) {
	return m.symbols, nil
}

func (m *mockHistoricRepo) GetHistoricSymbols() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.historicSymbols != nil {
		return m.historicSymbols, nil
	}
	seen := make(map[string]struct{})
	for _, v := range m.values {
		seen[v.Symbol] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockHistoricRepo) Insert(value *models.AssetHistoricValue) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values = append(m.values, *value)
	return nil
}

func (m *mockHistoricRepo) GetValues() []models.AssetHistoricValue {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.values
}

func TestHistoricPriceService_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	fetcher := pricesPkg.NewPriceService()
	repo := &mockHistoricRepo{}

	tests := []struct {
		name string
		opts []HistoricPriceOption
	}{
		{"no context", []HistoricPriceOption{
			WithHistoricPriceLogger(historicDiscardLogger),
			WithHistoricPriceFetcher(fetcher),
			WithHistoricPriceRepo(repo),
		}},
		{"no logger", []HistoricPriceOption{
			WithHistoricPriceContext(ctx),
			WithHistoricPriceFetcher(fetcher),
			WithHistoricPriceRepo(repo),
		}},
		{"no fetcher", []HistoricPriceOption{
			WithHistoricPriceContext(ctx),
			WithHistoricPriceLogger(historicDiscardLogger),
			WithHistoricPriceRepo(repo),
		}},
		{"no repo", []HistoricPriceOption{
			WithHistoricPriceContext(ctx),
			WithHistoricPriceLogger(historicDiscardLogger),
			WithHistoricPriceFetcher(fetcher),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHistoricPriceService(tt.opts...)
			assert.Error(t, err)
		})
	}
}

func TestHistoricPriceService_Tick(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := &mockHistoricRepo{
		symbols: []string{"BTC", "ETH"},
	}

	svc, err := NewHistoricPriceService(
		WithHistoricPriceContext(ctx),
		WithHistoricPriceLogger(historicDiscardLogger),
		WithHistoricPriceFetcher(fetcher),
		WithHistoricPriceRepo(repo),
	)
	require.NoError(t, err)

	err = svc.tick()
	require.NoError(t, err)

	values := repo.GetValues()
	assert.Len(t, values, 2)

	var btcValue, ethValue *models.AssetHistoricValue
	for i := range values {
		if values[i].Symbol == "BTC" {
			btcValue = &values[i]
		}
		if values[i].Symbol == "ETH" {
			ethValue = &values[i]
		}
	}

	require.NotNil(t, btcValue)
	require.NotNil(t, ethValue)
	assert.Greater(t, btcValue.Value, float64(0))
	assert.Greater(t, ethValue.Value, float64(0))
}

func TestHistoricPriceService_TickEmptyAssets(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	fetcher := pricesPkg.NewPriceService()
	repo := &mockHistoricRepo{
		symbols: []string{},
	}

	svc, err := NewHistoricPriceService(
		WithHistoricPriceContext(ctx),
		WithHistoricPriceLogger(historicDiscardLogger),
		WithHistoricPriceFetcher(fetcher),
		WithHistoricPriceRepo(repo),
	)
	require.NoError(t, err)

	err = svc.tick()
	require.NoError(t, err)

	values := repo.GetValues()
	assert.Len(t, values, 0)
}
