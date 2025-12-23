package repo

import (
	"hodlbook/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Asset{}, &models.Transaction{}, &models.Exchange{}, &models.Price{}))
	return db
}

func TestTransactionRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	tx := &models.Transaction{
		Type:      "deposit",
		AssetID:   1,
		Amount:    100.0,
		Notes:     "test deposit",
		Timestamp: time.Now(),
	}

	require.NoError(t, repository.CreateTransaction(tx))
	require.NotZero(t, tx.ID)

	got, err := repository.GetTransactionByID(tx.ID)
	require.NoError(t, err)
	require.Equal(t, tx.Amount, got.Amount)

	tx.Amount = 200.0
	require.NoError(t, repository.UpdateTransaction(tx))
	got, err = repository.GetTransactionByID(tx.ID)
	require.NoError(t, err)
	require.Equal(t, 200.0, got.Amount)

	transactions, err := repository.GetAllTransactions()
	require.NoError(t, err)
	require.Len(t, transactions, 1)

	require.NoError(t, repository.DeleteTransaction(tx.ID))
	_, err = repository.GetTransactionByID(tx.ID)
	require.Error(t, err)
}
