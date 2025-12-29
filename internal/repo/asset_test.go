package repo

import (
	"hodlbook/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	asset := &models.Asset{
		Symbol: "BTC",
		Name:   "Bitcoin",
	}

	require.NoError(t, repository.CreateAsset(asset))
	require.NotZero(t, asset.ID)

	got, err := repository.GetAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, asset.Symbol, got.Symbol)
	require.Equal(t, asset.Name, got.Name)

	assets, err := repository.GetAllAssets()
	require.NoError(t, err)
	require.Len(t, assets, 1)

	require.NoError(t, repository.DeleteAsset(asset.ID))
	_, err = repository.GetAssetByID(asset.ID)
	require.Error(t, err)
}

func TestAssetRepository_MultipleWithSameSymbol(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	asset1 := &models.Asset{Symbol: "ETH", Name: "Ethereum", TransactionType: "deposit", Amount: 1.0}
	asset2 := &models.Asset{Symbol: "ETH", Name: "Ethereum", TransactionType: "deposit", Amount: 2.0}

	require.NoError(t, repository.CreateAsset(asset1))
	require.NoError(t, repository.CreateAsset(asset2))

	assets, err := repository.GetAllAssets()
	require.NoError(t, err)
	require.Len(t, assets, 2)
}
