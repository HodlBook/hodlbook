package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPriceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Price{}))
	return db
}

func TestPriceRepository_CRUD(t *testing.T) {
	db := setupPriceTestDB(t)
	repo, err := NewPrice(WithPriceDB(db))
	require.NoError(t, err)

	price := &Price{
		AssetID:   1,
		Currency:  "USD",
		Price:     123.45,
		Timestamp: time.Now(),
	}

	// Create
	require.NoError(t, repo.Create(price))
	require.NotZero(t, price.ID)

	// GetByID
	got, err := repo.GetByID(price.ID)
	require.NoError(t, err)
	require.Equal(t, price.Price, got.Price)

	// Update
	price.Price = 200.0
	require.NoError(t, repo.Update(price))
	got, err = repo.GetByID(price.ID)
	require.NoError(t, err)
	require.Equal(t, 200.0, got.Price)

	// GetByAssetAndCurrency
	prices, err := repo.GetByAssetAndCurrency(1, "USD")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	// Delete
	require.NoError(t, repo.Delete(price.ID))
	_, err = repo.GetByID(price.ID)
	require.Error(t, err)
}
