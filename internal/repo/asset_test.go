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
		Symbol:   "BTC",
		Name:     "Bitcoin",
		Type:     "crypto",
		Decimals: 8,
	}

	require.NoError(t, repository.CreateAsset(asset))
	require.NotZero(t, asset.ID)

	got, err := repository.GetAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, asset.Symbol, got.Symbol)
	require.Equal(t, asset.Name, got.Name)

	got, err = repository.GetAssetBySymbol("BTC")
	require.NoError(t, err)
	require.Equal(t, asset.ID, got.ID)

	asset.Name = "Bitcoin (Updated)"
	require.NoError(t, repository.UpdateAsset(asset))
	got, err = repository.GetAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, "Bitcoin (Updated)", got.Name)

	assets, err := repository.GetAllAssets()
	require.NoError(t, err)
	require.Len(t, assets, 1)

	assets, err = repository.GetAssetsByType("crypto")
	require.NoError(t, err)
	require.Len(t, assets, 1)

	assets, err = repository.GetAssetsByType("fiat")
	require.NoError(t, err)
	require.Len(t, assets, 0)

	require.NoError(t, repository.DeleteAsset(asset.ID))
	_, err = repository.GetAssetByID(asset.ID)
	require.Error(t, err)
}

func TestAssetRepository_UniqueSymbol(t *testing.T) {
	db := setupTestDB(t)
	repository, err := New(db)
	require.NoError(t, err)

	asset1 := &models.Asset{Symbol: "ETH", Name: "Ethereum", Type: "crypto"}
	asset2 := &models.Asset{Symbol: "ETH", Name: "Ethereum Duplicate", Type: "crypto"}

	require.NoError(t, repository.CreateAsset(asset1))
	require.Error(t, repository.CreateAsset(asset2))
}
