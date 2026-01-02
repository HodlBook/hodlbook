package repo

import (
	"hodlbook/internal/models"
	"hodlbook/internal/repo"
)

type Repository interface {
	// Assets
	ListAssets(filter repo.AssetFilter) (*repo.AssetListResult, error)
	GetAssetByID(id int64) (*models.Asset, error)
	GetAllAssets() ([]models.Asset, error)
	GetAssetsBySymbol(symbol string) ([]models.Asset, error)
	CreateAsset(asset *models.Asset) error
	UpdateAsset(asset *models.Asset) error
	DeleteAsset(id int64) error
	GetUniqueSymbols() ([]string, error)

	// Exchanges
	ListExchanges(filter repo.ExchangeFilter) (*repo.ExchangeListResult, error)
	GetExchangeByID(id int64) (*models.Exchange, error)
	GetAllExchanges() ([]models.Exchange, error)
	CreateExchange(exchange *models.Exchange) error
	UpdateExchange(exchange *models.Exchange) error
	DeleteExchange(id int64) error

	// Price history
	SelectAllBySymbol(symbol string) ([]models.AssetHistoricValue, error)
	Insert(value *models.AssetHistoricValue) error

	// Prices
	CreatePrice(price *models.Price) error

	// Import logs
	CreateImportLog(log *models.ImportLog) error
	GetImportLogByID(id int64) (*models.ImportLog, error)
	ListImportLogs() ([]models.ImportLog, error)
	UpdateImportLog(log *models.ImportLog) error
	DeleteImportLog(id int64) error
}
