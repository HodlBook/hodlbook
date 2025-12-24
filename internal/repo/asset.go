package repo

import "hodlbook/internal/models"

func (r *Repository) CreateAsset(asset *models.Asset) error {
	return r.db.Create(asset).Error
}

func (r *Repository) GetAssetByID(id int64) (*models.Asset, error) {
	var asset models.Asset
	if err := r.db.First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *Repository) GetAllAssets() ([]models.Asset, error) {
	var assets []models.Asset
	if err := r.db.Order("symbol ASC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *Repository) UpdateAsset(asset *models.Asset) error {
	return r.db.Save(asset).Error
}

func (r *Repository) DeleteAsset(id int64) error {
	return r.db.Delete(&models.Asset{}, id).Error
}
