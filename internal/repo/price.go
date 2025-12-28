package repo

import (
	"hodlbook/internal/models"
	"time"
)

func (r *Repository) CreatePrice(price *models.Price) error {
	return r.db.Create(price).Error
}

func (r *Repository) GetPriceByID(id int64) (*models.Price, error) {
	var price models.Price
	if err := r.db.First(&price, id).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *Repository) GetLatestPrice(symbol, currency string) (*models.Price, error) {
	var price models.Price
	if err := r.db.Where("symbol = ? AND currency = ?", symbol, currency).
		Order("timestamp DESC").
		First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *Repository) GetPricesBySymbol(symbol string) ([]models.Price, error) {
	var prices []models.Price
	if err := r.db.Where("symbol = ?", symbol).Order("timestamp DESC").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *Repository) GetPricesBySymbolAndCurrency(symbol, currency string) ([]models.Price, error) {
	var prices []models.Price
	if err := r.db.Where("symbol = ? AND currency = ?", symbol, currency).
		Order("timestamp DESC").
		Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *Repository) GetPricesByDateRange(symbol, currency string, startDate, endDate time.Time) ([]models.Price, error) {
	var prices []models.Price
	if err := r.db.Where("symbol = ? AND currency = ? AND timestamp BETWEEN ? AND ?", symbol, currency, startDate, endDate).
		Order("timestamp ASC").
		Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *Repository) UpdatePrice(price *models.Price) error {
	return r.db.Save(price).Error
}

func (r *Repository) DeletePrice(id int64) error {
	return r.db.Delete(&models.Price{}, id).Error
}

func (r *Repository) DeletePricesOlderThan(date time.Time) error {
	return r.db.Where("timestamp < ?", date).Delete(&models.Price{}).Error
}
