package repo

import (
	"hodlbook/internal/models"
)

// Insert a new AssetHistoricValue
func (r *Repository) Insert(value *models.AssetHistoricValue) error {
	return r.db.Create(value).Error
}

// Select all AssetHistoricValues by AssetID
func (r *Repository) SelectAllByAsset(assetID int64) ([]models.AssetHistoricValue, error) {
	var values []models.AssetHistoricValue
	err := r.db.Where("asset_id = ?", assetID).Find(&values).Error
	return values, err
}
