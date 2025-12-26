package binanceprices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hodlbook/pkg/pairs"
	"hodlbook/pkg/types/prices"
)

var (
	_ prices.PriceFetcher = (*PriceFetcher)(nil)
)

type PriceFetcher struct {
	BaseURL string
	Client  *http.Client
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		BaseURL: "https://api.binance.com/api/v3",
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (b *PriceFetcher) Fetch(price *prices.Price) error {
	pair := price.Asset.Symbol + "USD" // Normalize to USD
	endpoint := fmt.Sprintf("%s/ticker/price?symbol=%s", b.BaseURL, pair)

	resp, err := b.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("invalid trading pair: %s", pair)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	priceValue, err := strconv.ParseFloat(result.Price, 64)
	if err != nil {
		return fmt.Errorf("invalid price format: %w", err)
	}

	price.Value = priceValue
	return nil
}

func (b *PriceFetcher) FetchMany(prices ...*prices.Price) error {
	endpoint := fmt.Sprintf("%s/ticker/price", b.BaseURL)

	resp, err := b.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []pairs.Pair
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	priceMap := pairs.PairMap(results)
	for _, price := range prices {
		pair := price.Asset.Symbol + "USD" // Normalize to USD
		priceValue, err := pairs.GetPriceForPair(pair, priceMap)
		if err != nil {
			return fmt.Errorf("failed to get price for pair %s: %w", pair, err)
		}
		price.Value = priceValue
	}

	return nil
}

func (b *PriceFetcher) FetchAll() ([]prices.Price, error) {
	endpoint := fmt.Sprintf("%s/ticker/price", b.BaseURL)

	resp, err := b.Client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	pricesList := make([]prices.Price, 0)
	for _, result := range results {
		if strings.HasSuffix(result.Symbol, "USD") {
			priceValue, err := strconv.ParseFloat(result.Price, 64)
			if err == nil {
				pricesList = append(pricesList, prices.Price{
					Asset: prices.Asset{Name: result.Symbol[:len(result.Symbol)-3], Symbol: result.Symbol[:len(result.Symbol)-3]},
					Value: priceValue,
				})
			}
		}
	}

	return pricesList, nil
}
