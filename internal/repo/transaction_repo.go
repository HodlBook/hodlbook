package repo

import (
	"time"

	"gorm.io/gorm"
)

// TransactionRepository handles transaction data operations
type TransactionRepository struct {
	db *gorm.DB
}

// Option is the functional options pattern for TransactionRepository
type TransactionOption func(*TransactionRepository) error

// NewTransaction creates a new transaction repository with options
func NewTransaction(opts ...TransactionOption) (*TransactionRepository, error) {
	r := &TransactionRepository{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// WithDB sets the database instance for TransactionRepository
func WithDB(db *gorm.DB) TransactionOption {
	return func(r *TransactionRepository) error {
		if db == nil {
			return ErrNilDatabase
		}
		r.db = db
		return nil
	}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(tx *Transaction) error {
	return r.db.Create(tx).Error
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(id int64) (*Transaction, error) {
	var tx Transaction
	if err := r.db.First(&tx, id).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetAll retrieves all transactions
func (r *TransactionRepository) GetAll() ([]Transaction, error) {
	var transactions []Transaction
	if err := r.db.Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetByAssetID retrieves transactions for a specific asset
func (r *TransactionRepository) GetByAssetID(assetID int64) ([]Transaction, error) {
	var transactions []Transaction
	if err := r.db.Where("asset_id = ?", assetID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetByType retrieves transactions by type (deposit/withdrawal)
func (r *TransactionRepository) GetByType(txType string) ([]Transaction, error) {
	var transactions []Transaction
	if err := r.db.Where("type = ?", txType).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// Update updates an existing transaction
func (r *TransactionRepository) Update(tx *Transaction) error {
	return r.db.Save(tx).Error
}

// Delete deletes a transaction
func (r *TransactionRepository) Delete(id int64) error {
	return r.db.Delete(&Transaction{}, id).Error
}

// GetByDateRange retrieves transactions within a date range
func (r *TransactionRepository) GetByDateRange(startDate, endDate time.Time) ([]Transaction, error) {
	var transactions []Transaction
	if err := r.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate).Order("timestamp DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTotalByAssetAndType calculates total amount for an asset by type
func (r *TransactionRepository) GetTotalByAssetAndType(assetID int64, txType string) (float64, error) {
	var total float64
	if err := r.db.Model(&Transaction{}).Where("asset_id = ? AND type = ?", assetID, txType).Pluck("COALESCE(SUM(amount), 0)", &total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
