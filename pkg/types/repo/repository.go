package repo

import (
	"hodlbook/internal/models"
	"hodlbook/internal/repo"
)

type Repository interface {
	// Assets
	GetAllAssets() ([]models.Asset, error)
	GetAssetByID(id int64) (*models.Asset, error)
	CreateAsset(asset *models.Asset) error
	DeleteAsset(id int64) error

	// Transactions
	ListTransactions(filter repo.TransactionFilter) (*repo.TransactionListResult, error)
	GetTransactionByID(id int64) (*models.Transaction, error)
	CreateTransaction(tx *models.Transaction) error
	UpdateTransaction(tx *models.Transaction) error
	DeleteTransaction(id int64) error

	// Exchanges
	ListExchanges(filter repo.ExchangeFilter) (*repo.ExchangeListResult, error)
	GetExchangeByID(id int64) (*models.Exchange, error)
	CreateExchange(exchange *models.Exchange) error
	UpdateExchange(exchange *models.Exchange) error
	DeleteExchange(id int64) error

	// Price history
	SelectAllByAsset(assetID int64) ([]models.AssetHistoricValue, error)
}
