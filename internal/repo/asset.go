package repo

import (
	"hodlbook/internal/models"
	"time"
)

type AssetFilter struct {
	Symbol          string
	TransactionType string
	StartDate       *time.Time
	EndDate         *time.Time
	Limit           int
	Offset          int
}

type AssetListResult struct {
	Assets []models.Asset `json:"assets"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

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
	if err := r.db.Order("timestamp DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *Repository) GetAssetsBySymbol(symbol string) ([]models.Asset, error) {
	var assets []models.Asset
	if err := r.db.Where("symbol = ?", symbol).Order("timestamp DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *Repository) GetAssetsByType(txType string) ([]models.Asset, error) {
	var assets []models.Asset
	if err := r.db.Where("transaction_type = ?", txType).Order("timestamp DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *Repository) UpdateAsset(asset *models.Asset) error {
	return r.db.Model(asset).Updates(asset).Error
}

func (r *Repository) DeleteAsset(id int64) error {
	return r.db.Delete(&models.Asset{}, id).Error
}

func (r *Repository) GetAssetsByDateRange(startDate, endDate time.Time) ([]models.Asset, error) {
	var assets []models.Asset
	if err := r.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate).Order("timestamp DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *Repository) GetTotalBySymbolAndType(symbol string, txType string) (float64, error) {
	var total float64
	if err := r.db.Model(&models.Asset{}).Where("symbol = ? AND transaction_type = ?", symbol, txType).Pluck("COALESCE(SUM(amount), 0)", &total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) ListAssets(filter AssetFilter) (*AssetListResult, error) {
	query := r.db.Model(&models.Asset{})

	if filter.Symbol != "" {
		query = query.Where("symbol = ?", filter.Symbol)
	}
	if filter.TransactionType != "" {
		query = query.Where("transaction_type = ?", filter.TransactionType)
	}
	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", *filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var assets []models.Asset
	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&assets).Error; err != nil {
		return nil, err
	}

	return &AssetListResult{
		Assets: assets,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *Repository) CountAssets() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Asset{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetUniqueSymbols() ([]string, error) {
	var symbols []string
	if err := r.db.Model(&models.Asset{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return nil, err
	}
	return symbols, nil
}

func (r *Repository) GetFirstAssetBySymbol(symbol string) (*models.Asset, error) {
	var asset models.Asset
	if err := r.db.Where("symbol = ?", symbol).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}
