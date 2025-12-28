package repo

import (
	"hodlbook/internal/models"
)

func (r *Repository) Insert(value *models.AssetHistoricValue) error {
	return r.db.Create(value).Error
}

func (r *Repository) SelectAllBySymbol(symbol string) ([]models.AssetHistoricValue, error) {
	var values []models.AssetHistoricValue
	err := r.db.Where("symbol = ?", symbol).Order("timestamp DESC").Find(&values).Error
	return values, err
}
