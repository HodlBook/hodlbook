package models

import "time"

type Asset struct {
	ID              int64     `json:"id"               gorm:"primaryKey"`
	Symbol          string    `json:"symbol"           gorm:"index"`
	Name            string    `json:"name"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type" gorm:"index"`
	Notes           string    `json:"notes"`
	PriceSource     *string   `json:"price_source,omitempty"`
	Timestamp       time.Time `json:"timestamp"        gorm:"index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AssetHistoricValue struct {
	Symbol    string    `json:"symbol"     gorm:"index"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"  gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
}

type Exchange struct {
	ID          int64     `json:"id"          gorm:"primaryKey"`
	FromSymbol  string    `json:"from_symbol" gorm:"index"`
	ToSymbol    string    `json:"to_symbol"   gorm:"index"`
	FromAmount  float64   `json:"from_amount"`
	ToAmount    float64   `json:"to_amount"`
	Fee         float64   `json:"fee"`
	FeeCurrency string    `json:"fee_currency"`
	Notes       string    `json:"notes"`
	Timestamp   time.Time `json:"timestamp"   gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Price struct {
	ID        int64     `json:"id"         gorm:"primaryKey"`
	Symbol    string    `json:"symbol"     gorm:"index:idx_symbol_currency_time"`
	Currency  string    `json:"currency"   gorm:"index:idx_symbol_currency_time"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"  gorm:"index:idx_symbol_currency_time"`
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	ID    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ImportLog struct {
	ID           int64     `json:"id"            gorm:"primaryKey"`
	Filename     string    `json:"filename"`
	Format       string    `json:"format"`
	EntityType   string    `json:"entity_type"`
	TotalRows    int       `json:"total_rows"`
	ImportedRows int       `json:"imported_rows"`
	FailedRows   int       `json:"failed_rows"`
	Status       string    `json:"status"`
	FailedData   string    `json:"failed_data"   gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Asset) TableName() string {
	return "assets"
}

func (AssetHistoricValue) TableName() string {
	return "asset_historic_values"
}

func (Exchange) TableName() string {
	return "exchanges"
}

func (Price) TableName() string {
	return "prices"
}

func (ImportLog) TableName() string {
	return "import_logs"
}
