package repo

import "time"

// Transaction represents a deposit or withdrawal transaction
type Transaction struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"index" json:"type"` // "deposit" or "withdrawal"
	AssetID   int64     `gorm:"index" json:"asset_id"`
	Amount    float64   `json:"amount"`
	Notes     string    `json:"notes"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Price represents the historic price of an asset
type Price struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AssetID   int64     `gorm:"index:idx_asset_currency_time" json:"asset_id"`
	Currency  string    `gorm:"index:idx_asset_currency_time" json:"currency"`
	Price     float64   `json:"price"`
	Timestamp time.Time `gorm:"index:idx_asset_currency_time" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// TableName specifies the table name for Price
func (Price) TableName() string {
	return "prices"
}
