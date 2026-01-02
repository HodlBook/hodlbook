package defillamaprices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hodlbook/pkg/types/prices"
)

var (
	_ prices.PriceFetcher = (*PriceFetcher)(nil)
)

type PriceFetcher struct {
	BaseURL  string
	Client   *http.Client
	Network  string
	Contract string
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		BaseURL: "https://coins.llama.fi",
		Client:  &http.Client{Timeout: 10 * time.Second},
		Network: "ethereum",
	}
}

func NewPriceFetcherForToken(network, contract string) *PriceFetcher {
	f := NewPriceFetcher()
	f.Network = network
	f.Contract = contract
	return f
}

func (d *PriceFetcher) Fetch(price *prices.Price) error {
	symbol := strings.ToUpper(price.Asset.Symbol)
	if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
		price.Value = 1.0
		return nil
	}

	var identifier string
	if d.Contract != "" {
		identifier = fmt.Sprintf("%s:%s", d.Network, d.Contract)
	} else if isContractAddress(price.Asset.Name) {
		identifier = fmt.Sprintf("%s:%s", d.Network, price.Asset.Name)
	} else {
		identifier = fmt.Sprintf("coingecko:%s", strings.ToLower(price.Asset.Name))
	}

	endpoint := fmt.Sprintf("%s/prices/current/%s", d.BaseURL, identifier)

	resp, err := d.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Coins map[string]struct {
			Price     float64 `json:"price"`
			Symbol    string  `json:"symbol"`
			Timestamp int64   `json:"timestamp"`
		} `json:"coins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for id, coin := range result.Coins {
		price.Value = coin.Price
		if strings.Contains(id, ":0x") {
			parts := strings.SplitN(id, ":", 2)
			if len(parts) == 2 {
				price.Network = parts[0]
				price.PoolAddress = parts[1]
			}
		}
		return nil
	}

	return fmt.Errorf("price not found for %s", symbol)
}

func isContractAddress(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "0x") && len(s) == 42
}

func (d *PriceFetcher) FetchMany(pairs ...*prices.Price) error {
	ids := make([]string, 0, len(pairs))
	idToPair := make(map[string]*prices.Price)

	for _, pair := range pairs {
		symbol := strings.ToUpper(pair.Asset.Symbol)
		if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
			pair.Value = 1.0
			continue
		}
		id := fmt.Sprintf("coingecko:%s", strings.ToLower(pair.Asset.Name))
		ids = append(ids, id)
		idToPair[id] = pair
	}

	if len(ids) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/prices/current/%s", d.BaseURL, strings.Join(ids, ","))

	resp, err := d.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Coins map[string]struct {
			Price     float64 `json:"price"`
			Symbol    string  `json:"symbol"`
			Timestamp int64   `json:"timestamp"`
		} `json:"coins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for id, coin := range result.Coins {
		if pair, ok := idToPair[id]; ok {
			pair.Value = coin.Price
		}
	}

	return nil
}

func (d *PriceFetcher) FetchAll() ([]prices.Price, error) {
	return nil, fmt.Errorf("FetchAll not supported: DefiLlama requires specific coin identifiers")
}
