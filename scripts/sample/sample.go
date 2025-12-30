package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const baseURL = "http://localhost:2008/api"

type Asset struct {
	ID     int64  `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type Transaction struct {
	Type      string    `json:"type"`
	AssetID   int64     `json:"asset_id"`
	Amount    float64   `json:"amount"`
	Notes     string    `json:"notes"`
	Timestamp time.Time `json:"timestamp"`
}

type Exchange struct {
	FromAssetID int64     `json:"from_asset_id"`
	ToAssetID   int64     `json:"to_asset_id"`
	FromAmount  float64   `json:"from_amount"`
	ToAmount    float64   `json:"to_amount"`
	Notes       string    `json:"notes"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	usd := createAsset("USD", "US Dollar")
	btc := createAsset("BTC", "Bitcoin")
	eth := createAsset("ETH", "Ethereum")

	fmt.Printf("Created assets: USD=%d, BTC=%d, ETH=%d\n", usd.ID, btc.ID, eth.ID)

	startDate := time.Now().AddDate(0, 0, -7)
	createTransaction(Transaction{
		Type:      "deposit",
		AssetID:   usd.ID,
		Amount:    10000,
		Notes:     "Initial USD deposit",
		Timestamp: startDate,
	})
	fmt.Println("Deposited 10000 USD")

	usdBalance := 10000.0
	btcPrice := 95000.0
	ethPrice := 3400.0

	for day := 0; day < 7; day++ {
		txDate := startDate.AddDate(0, 0, day+1)

		btcBuy := usdBalance * 0.01
		btcAmount := btcBuy / btcPrice
		createExchange(Exchange{
			FromAssetID: usd.ID,
			ToAssetID:   btc.ID,
			FromAmount:  btcBuy,
			ToAmount:    btcAmount,
			Notes:       fmt.Sprintf("Day %d: Buy BTC", day+1),
			Timestamp:   txDate,
		})
		usdBalance -= btcBuy
		fmt.Printf("Day %d: Bought %.8f BTC for %.2f USD\n", day+1, btcAmount, btcBuy)

		ethBuy := usdBalance * 0.01
		ethAmount := ethBuy / ethPrice
		createExchange(Exchange{
			FromAssetID: usd.ID,
			ToAssetID:   eth.ID,
			FromAmount:  ethBuy,
			ToAmount:    ethAmount,
			Notes:       fmt.Sprintf("Day %d: Buy ETH", day+1),
			Timestamp:   txDate,
		})
		usdBalance -= ethBuy
		fmt.Printf("Day %d: Bought %.8f ETH for %.2f USD\n", day+1, ethAmount, ethBuy)
	}

	fmt.Printf("\nFinal USD balance: %.2f\n", usdBalance)
	fmt.Println("Sample data created successfully!")
}

func createAsset(symbol, name string) Asset {
	asset := Asset{Symbol: symbol, Name: name}
	body, _ := json.Marshal(asset)

	resp, err := http.Post(baseURL+"/assets", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to create asset %s: %v", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create asset %s: status %d", symbol, resp.StatusCode)
	}

	var created Asset
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		log.Fatalf("Failed to decode asset response: %v", err)
	}
	return created
}

func createTransaction(tx Transaction) {
	body, _ := json.Marshal(tx)

	resp, err := http.Post(baseURL+"/transactions", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create transaction: status %d", resp.StatusCode)
	}
}

func createExchange(ex Exchange) {
	body, _ := json.Marshal(ex)

	resp, err := http.Post(baseURL+"/exchanges", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to create exchange: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create exchange: status %d", resp.StatusCode)
	}
}
