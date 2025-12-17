package repo

import (
	"time"

	"gorm.io/gorm"
)

// PriceRepository handles historic price data operations
type PriceRepository struct {
	db *gorm.DB
}

// Option is the functional options pattern for PriceRepository
type PriceOption func(*PriceRepository) error

// NewPrice creates a new price repository with options
func NewPrice(opts ...PriceOption) (*PriceRepository, error) {
	r := &PriceRepository{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// WithDB sets the database instance for PriceRepository
func WithPriceDB(db *gorm.DB) PriceOption {
	return func(r *PriceRepository) error {
		if db == nil {
			return ErrNilDatabase
		}
		r.db = db
		return nil
	}
}

// Create creates a new price record
func (r *PriceRepository) Create(price *Price) error {
	return r.db.Create(price).Error
}

// GetByID retrieves a price record by ID
func (r *PriceRepository) GetByID(id int64) (*Price, error) {
	var price Price
	if err := r.db.First(&price, id).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

// GetLatestByAssetAndCurrency gets the latest price for an asset in a specific currency
func (r *PriceRepository) GetLatestByAssetAndCurrency(assetID int64, currency string) (*Price, error) {
	var price Price
	if err := r.db.Where("asset_id = ? AND currency = ?", assetID, currency).
		Order("timestamp DESC").
		First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

// GetByAssetID retrieves all prices for an asset
func (r *PriceRepository) GetByAssetID(assetID int64) ([]Price, error) {
	var prices []Price
	if err := r.db.Where("asset_id = ?", assetID).Order("timestamp DESC").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

// GetByAssetAndCurrency retrieves prices for an asset in a specific currency
func (r *PriceRepository) GetByAssetAndCurrency(assetID int64, currency string) ([]Price, error) {
	var prices []Price
	if err := r.db.Where("asset_id = ? AND currency = ?", assetID, currency).
		Order("timestamp DESC").
		Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

// GetByDateRange retrieves price records within a date range
func (r *PriceRepository) GetByDateRange(assetID int64, currency string, startDate, endDate time.Time) ([]Price, error) {
	var prices []Price
	if err := r.db.Where("asset_id = ? AND currency = ? AND timestamp BETWEEN ? AND ?", assetID, currency, startDate, endDate).
		Order("timestamp ASC").
		Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

// Update updates a price record
func (r *PriceRepository) Update(price *Price) error {
	return r.db.Save(price).Error
}

// Delete deletes a price record
func (r *PriceRepository) Delete(id int64) error {
	return r.db.Delete(&Price{}, id).Error
}

// DeleteOlderThan deletes price records older than a specific date (for cleanup)
func (r *PriceRepository) DeleteOlderThan(date time.Time) error {
	return r.db.Where("timestamp < ?", date).Delete(&Price{}).Error
}
