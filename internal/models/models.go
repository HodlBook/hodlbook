package models

import "time"

type Asset struct {
	ID        int64     `json:"id"         gorm:"primaryKey"`
	Symbol    string    `json:"symbol"     gorm:"uniqueIndex"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AssetHistoricValue struct {
	AssetID   int64     `json:"asset_id"   gorm:"index"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"  gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID        int64     `json:"id"         gorm:"primaryKey"`
	Type      string    `json:"type"       gorm:"index"`
	AssetID   int64     `json:"asset_id"   gorm:"index"`
	Amount    float64   `json:"amount"`
	Notes     string    `json:"notes"`
	Timestamp time.Time `json:"timestamp"  gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Exchange struct {
	ID          int64     `json:"id"            gorm:"primaryKey"`
	FromAssetID int64     `json:"from_asset_id" gorm:"index"`
	ToAssetID   int64     `json:"to_asset_id"   gorm:"index"`
	FromAmount  float64   `json:"from_amount"`
	ToAmount    float64   `json:"to_amount"`
	Fee         float64   `json:"fee"`
	FeeCurrency string    `json:"fee_currency"`
	Notes       string    `json:"notes"`
	Timestamp   time.Time `json:"timestamp"     gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Price struct {
	ID        int64     `json:"id"         gorm:"primaryKey"`
	AssetID   int64     `json:"asset_id"   gorm:"index:idx_asset_currency_time"`
	Currency  string    `json:"currency"   gorm:"index:idx_asset_currency_time"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"  gorm:"index:idx_asset_currency_time"`
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	ID    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (Asset) TableName() string {
	return "assets"
}

func (Transaction) TableName() string {
	return "transactions"
}

func (Exchange) TableName() string {
	return "exchanges"
}

func (Price) TableName() string {
	return "prices"
}
