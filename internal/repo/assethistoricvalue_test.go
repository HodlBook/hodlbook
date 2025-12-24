package repo

import (
	"testing"
	"time"

	"hodlbook/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAssetHistoricValues(t *testing.T) {
	db := setupTestDB(t)

	// Auto-migrate the schema
	db.AutoMigrate(&models.AssetHistoricValue{})

	repo := &Repository{db: db}

	// Insert a new AssetHistoricValue
	historicValue := &models.AssetHistoricValue{
		AssetID:   1,
		Value:     123.45,
		Timestamp: time.Now(),
	}
	err := repo.Insert(historicValue)
	assert.NoError(t, err)

	// Select all AssetHistoricValues by AssetID
	values, err := repo.SelectAllByAsset(1)
	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, historicValue.Value, values[0].Value)
	assert.Equal(t, historicValue.AssetID, values[0].AssetID)
}
