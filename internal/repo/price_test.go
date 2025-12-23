package repo

import (
	"hodlbook/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPriceRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	price := &models.Price{
		AssetID:   1,
		Currency:  "USD",
		Price:     123.45,
		Timestamp: time.Now(),
	}

	require.NoError(t, repository.CreatePrice(price))
	require.NotZero(t, price.ID)

	got, err := repository.GetPriceByID(price.ID)
	require.NoError(t, err)
	require.Equal(t, price.Price, got.Price)

	price.Price = 200.0
	require.NoError(t, repository.UpdatePrice(price))
	got, err = repository.GetPriceByID(price.ID)
	require.NoError(t, err)
	require.Equal(t, 200.0, got.Price)

	prices, err := repository.GetPricesByAssetAndCurrency(1, "USD")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	require.NoError(t, repository.DeletePrice(price.ID))
	_, err = repository.GetPriceByID(price.ID)
	require.Error(t, err)
}
