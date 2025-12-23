package repo

import (
	"hodlbook/internal/models"
	"time"
)

type TransactionFilter struct {
	AssetID   *int64
	Type      string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

type TransactionListResult struct {
	Transactions []models.Transaction `json:"transactions"`
	Total        int64                `json:"total"`
	Limit        int                  `json:"limit"`
	Offset       int                  `json:"offset"`
}

func (r *Repository) CreateTransaction(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *Repository) GetTransactionByID(id int64) (*models.Transaction, error) {
	var tx models.Transaction
	if err := r.db.First(&tx, id).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *Repository) GetAllTransactions() ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Order("timestamp DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *Repository) GetTransactionsByAssetID(assetID int64) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Where("asset_id = ?", assetID).Order("timestamp DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *Repository) GetTransactionsByType(txType string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Where("type = ?", txType).Order("timestamp DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *Repository) UpdateTransaction(tx *models.Transaction) error {
	return r.db.Save(tx).Error
}

func (r *Repository) DeleteTransaction(id int64) error {
	return r.db.Delete(&models.Transaction{}, id).Error
}

func (r *Repository) GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate).Order("timestamp DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *Repository) GetTotalByAssetAndType(assetID int64, txType string) (float64, error) {
	var total float64
	if err := r.db.Model(&models.Transaction{}).Where("asset_id = ? AND type = ?", assetID, txType).Pluck("COALESCE(SUM(amount), 0)", &total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) ListTransactions(filter TransactionFilter) (*TransactionListResult, error) {
	query := r.db.Model(&models.Transaction{})

	if filter.AssetID != nil {
		query = query.Where("asset_id = ?", *filter.AssetID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
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

	var transactions []models.Transaction
	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &TransactionListResult{
		Transactions: transactions,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}, nil
}

func (r *Repository) CountTransactions() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
