package repo

import (
	"hodlbook/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Asset{},
		&models.AssetHistoricValue{},
		&models.Exchange{},
		&models.Price{},
	))
	return db
}
