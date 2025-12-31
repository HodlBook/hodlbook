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

func (r *Repository) GetPriceAtTime(symbol, currency string, timestamp time.Time) (*models.Price, error) {
	var prices []models.Price
	if err := r.db.Where("symbol = ? AND currency = ? AND timestamp <= ?", symbol, currency, timestamp).
		Order("timestamp DESC").
		Limit(1).
		Find(&prices).Error; err != nil {
		return nil, err
	}
	if len(prices) == 0 {
		return nil, nil
	}
	return &prices[0], nil
}

type AssetPriceQuery struct {
	Symbol    string
	Timestamp time.Time
}

func (r *Repository) GetPricesAtTimes(queries []AssetPriceQuery, currency string) (map[string]map[int64]float64, error) {
	result := make(map[string]map[int64]float64)

	for _, q := range queries {
		price, err := r.GetPriceAtTime(q.Symbol, currency, q.Timestamp)
		if err != nil {
			continue
		}
		if price == nil {
			continue
		}

		if result[q.Symbol] == nil {
			result[q.Symbol] = make(map[int64]float64)
		}
		result[q.Symbol][q.Timestamp.Unix()] = price.Price
	}

	return result, nil
}
