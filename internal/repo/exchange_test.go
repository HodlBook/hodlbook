package repo

import (
	"hodlbook/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExchangeRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	exchange := &models.Exchange{
		FromSymbol:  "BTC",
		ToSymbol:    "ETH",
		FromAmount:  1.0,
		ToAmount:    30000.0,
		Fee:         0.001,
		FeeCurrency: "BTC",
		Notes:       "test exchange",
		Timestamp:   time.Now(),
	}

	require.NoError(t, repository.CreateExchange(exchange))
	require.NotZero(t, exchange.ID)

	got, err := repository.GetExchangeByID(exchange.ID)
	require.NoError(t, err)
	require.Equal(t, exchange.FromAmount, got.FromAmount)
	require.Equal(t, exchange.ToAmount, got.ToAmount)

	exchange.ToAmount = 31000.0
	require.NoError(t, repository.UpdateExchange(exchange))
	got, err = repository.GetExchangeByID(exchange.ID)
	require.NoError(t, err)
	require.Equal(t, 31000.0, got.ToAmount)

	exchanges, err := repository.GetAllExchanges()
	require.NoError(t, err)
	require.Len(t, exchanges, 1)

	require.NoError(t, repository.DeleteExchange(exchange.ID))
	_, err = repository.GetExchangeByID(exchange.ID)
	require.Error(t, err)
}

func TestExchangeRepository_GetBySymbol(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	exchange1 := &models.Exchange{FromSymbol: "BTC", ToSymbol: "ETH", FromAmount: 1.0, ToAmount: 15.0, Timestamp: time.Now()}
	exchange2 := &models.Exchange{FromSymbol: "ETH", ToSymbol: "SOL", FromAmount: 15.0, ToAmount: 30000.0, Timestamp: time.Now()}

	require.NoError(t, repository.CreateExchange(exchange1))
	require.NoError(t, repository.CreateExchange(exchange2))

	exchanges, err := repository.GetExchangesBySymbol("ETH")
	require.NoError(t, err)
	require.Len(t, exchanges, 2)

	exchanges, err = repository.GetExchangesByFromSymbol("BTC")
	require.NoError(t, err)
	require.Len(t, exchanges, 1)

	exchanges, err = repository.GetExchangesByToSymbol("SOL")
	require.NoError(t, err)
	require.Len(t, exchanges, 1)
}

func TestExchangeRepository_GetByDateRange(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	exchange1 := &models.Exchange{FromSymbol: "BTC", ToSymbol: "ETH", FromAmount: 1.0, ToAmount: 15.0, Timestamp: twoDaysAgo}
	exchange2 := &models.Exchange{FromSymbol: "BTC", ToSymbol: "ETH", FromAmount: 2.0, ToAmount: 30.0, Timestamp: now}

	require.NoError(t, repository.CreateExchange(exchange1))
	require.NoError(t, repository.CreateExchange(exchange2))

	exchanges, err := repository.GetExchangesByDateRange(yesterday, now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, exchanges, 1)
	require.Equal(t, 2.0, exchanges[0].FromAmount)
}
