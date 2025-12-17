package models

import "time"

// Asset represents a cryptocurrency or token
type Asset struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transaction represents a deposit or withdrawal
type Transaction struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"` // "deposit" or "withdrawal"
	AssetID   int64     `json:"asset_id"`
	Amount    float64   `json:"amount"`
	Notes     string    `json:"notes"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Exchange represents a swap from one asset to another
type Exchange struct {
	ID          int64     `json:"id"`
	FromAssetID int64     `json:"from_asset_id"`
	ToAssetID   int64     `json:"to_asset_id"`
	FromAmount  float64   `json:"from_amount"`
	ToAmount    float64   `json:"to_amount"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Price represents the price of an asset at a specific time
type Price struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	Currency  string    `json:"currency"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// Setting represents application configuration
type Setting struct {
	ID    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
