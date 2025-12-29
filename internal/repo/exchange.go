package repo

import (
	"hodlbook/internal/models"
	"time"
)

type ExchangeFilter struct {
	FromSymbol *string
	ToSymbol   *string
	Symbol     *string
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

type ExchangeListResult struct {
	Exchanges []models.Exchange `json:"exchanges"`
	Total     int64             `json:"total"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
}

func (r *Repository) CreateExchange(exchange *models.Exchange) error {
	return r.db.Create(exchange).Error
}

func (r *Repository) GetExchangeByID(id int64) (*models.Exchange, error) {
	var exchange models.Exchange
	if err := r.db.First(&exchange, id).Error; err != nil {
		return nil, err
	}
	return &exchange, nil
}

func (r *Repository) GetAllExchanges() ([]models.Exchange, error) {
	var exchanges []models.Exchange
	if err := r.db.Order("timestamp DESC").Find(&exchanges).Error; err != nil {
		return nil, err
	}
	return exchanges, nil
}

func (r *Repository) GetExchangesBySymbol(symbol string) ([]models.Exchange, error) {
	var exchanges []models.Exchange
	if err := r.db.Where("from_symbol = ? OR to_symbol = ?", symbol, symbol).Order("timestamp DESC").Find(&exchanges).Error; err != nil {
		return nil, err
	}
	return exchanges, nil
}

func (r *Repository) GetExchangesByFromSymbol(symbol string) ([]models.Exchange, error) {
	var exchanges []models.Exchange
	if err := r.db.Where("from_symbol = ?", symbol).Order("timestamp DESC").Find(&exchanges).Error; err != nil {
		return nil, err
	}
	return exchanges, nil
}

func (r *Repository) GetExchangesByToSymbol(symbol string) ([]models.Exchange, error) {
	var exchanges []models.Exchange
	if err := r.db.Where("to_symbol = ?", symbol).Order("timestamp DESC").Find(&exchanges).Error; err != nil {
		return nil, err
	}
	return exchanges, nil
}

func (r *Repository) GetExchangesByDateRange(startDate, endDate time.Time) ([]models.Exchange, error) {
	var exchanges []models.Exchange
	if err := r.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate).Order("timestamp DESC").Find(&exchanges).Error; err != nil {
		return nil, err
	}
	return exchanges, nil
}

func (r *Repository) UpdateExchange(exchange *models.Exchange) error {
	return r.db.Save(exchange).Error
}

func (r *Repository) DeleteExchange(id int64) error {
	return r.db.Delete(&models.Exchange{}, id).Error
}

func (r *Repository) GetUniqueExchangeSymbols() ([]string, error) {
	var fromSymbols, toSymbols []string

	if err := r.db.Model(&models.Exchange{}).Distinct("from_symbol").Pluck("from_symbol", &fromSymbols).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.Exchange{}).Distinct("to_symbol").Pluck("to_symbol", &toSymbols).Error; err != nil {
		return nil, err
	}

	symbolSet := make(map[string]bool)
	for _, s := range fromSymbols {
		symbolSet[s] = true
	}
	for _, s := range toSymbols {
		symbolSet[s] = true
	}

	symbols := make([]string, 0, len(symbolSet))
	for s := range symbolSet {
		symbols = append(symbols, s)
	}
	return symbols, nil
}

func (r *Repository) ListExchanges(filter ExchangeFilter) (*ExchangeListResult, error) {
	query := r.db.Model(&models.Exchange{})

	if filter.Symbol != nil {
		query = query.Where("from_symbol = ? OR to_symbol = ?", *filter.Symbol, *filter.Symbol)
	} else {
		if filter.FromSymbol != nil {
			query = query.Where("from_symbol = ?", *filter.FromSymbol)
		}
		if filter.ToSymbol != nil {
			query = query.Where("to_symbol = ?", *filter.ToSymbol)
		}
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

	var exchanges []models.Exchange
	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&exchanges).Error; err != nil {
		return nil, err
	}

	return &ExchangeListResult{
		Exchanges: exchanges,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}
