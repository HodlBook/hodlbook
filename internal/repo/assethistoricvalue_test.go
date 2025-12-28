package repo

import (
	"testing"
	"time"

	"hodlbook/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAssetHistoricValues(t *testing.T) {
	db := setupTestDB(t)

	db.AutoMigrate(&models.AssetHistoricValue{})

	repo := &Repository{db: db}

	historicValue := &models.AssetHistoricValue{
		Symbol:    "BTC",
		Value:     123.45,
		Timestamp: time.Now(),
	}
	err := repo.Insert(historicValue)
	assert.NoError(t, err)

	values, err := repo.SelectAllBySymbol("BTC")
	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, historicValue.Value, values[0].Value)
	assert.Equal(t, historicValue.Symbol, values[0].Symbol)
}
