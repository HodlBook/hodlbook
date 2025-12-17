package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Transaction{}, &Price{}))
	return db
}
func TestTransactionRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewTransaction(WithDB(db))
	require.NoError(t, err)

	tx := &Transaction{
		Type:      "deposit",
		AssetID:   1,
		Amount:    100.0,
		Notes:     "test deposit",
		Timestamp: time.Now(),
	}

	// Create
	require.NoError(t, repo.Create(tx))
	require.NotZero(t, tx.ID)

	// GetByID
	got, err := repo.GetByID(tx.ID)
	require.NoError(t, err)
	require.Equal(t, tx.Amount, got.Amount)

	// Update
	tx.Amount = 200.0
	require.NoError(t, repo.Update(tx))
	got, err = repo.GetByID(tx.ID)
	require.NoError(t, err)
	require.Equal(t, 200.0, got.Amount)

	// GetAll
	transactions, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, transactions, 1)

	// Delete
	require.NoError(t, repo.Delete(tx.ID))
	_, err = repo.GetByID(tx.ID)
	require.Error(t, err)
}
