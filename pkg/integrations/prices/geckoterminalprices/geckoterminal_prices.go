package geckoterminalprices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hodlbook/pkg/types/prices"
)

var (
	_ prices.PriceFetcher = (*PriceFetcher)(nil)
)

type PriceFetcher struct {
	BaseURL string
	Client  *http.Client
	Network string
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		BaseURL: "https://api.geckoterminal.com/api/v2",
		Client:  &http.Client{Timeout: 10 * time.Second},
		Network: "eth",
	}
}

func NewPriceFetcherForNetwork(network string) *PriceFetcher {
	f := NewPriceFetcher()
	f.Network = network
	return f
}

func (g *PriceFetcher) Fetch(price *prices.Price) error {
	symbol := strings.ToUpper(price.Asset.Symbol)
	if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
		price.Value = 1.0
		return nil
	}

	if isPoolAddress(price.Asset.Name) || isPoolAddress(price.Asset.Symbol) {
		return g.fetchByPoolAddress(price)
	}

	return g.fetchBySearch(price)
}

func isPoolAddress(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "0x") && len(s) == 42
}

func (g *PriceFetcher) fetchByPoolAddress(price *prices.Price) error {
	poolAddress := price.Asset.Name
	if isPoolAddress(price.Asset.Symbol) {
		poolAddress = price.Asset.Symbol
	}

	endpoint := fmt.Sprintf("%s/networks/%s/pools/%s", g.BaseURL, g.Network, strings.ToLower(poolAddress))

	resp, err := g.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch pool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Attributes struct {
				BaseTokenPriceUSD string `json:"base_token_price_usd"`
				Name              string `json:"name"`
				Address           string `json:"address"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	priceStr := result.Data.Attributes.BaseTokenPriceUSD
	if priceStr == "" {
		return fmt.Errorf("no price available for pool %s", poolAddress)
	}

	priceValue, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return fmt.Errorf("invalid price format: %w", err)
	}

	price.Value = priceValue
	price.PoolAddress = result.Data.Attributes.Address
	price.Network = g.Network
	price.Asset.Symbol = extractSymbol(result.Data.Attributes.Name)
	price.Asset.Name = result.Data.Attributes.Name
	return nil
}

func (g *PriceFetcher) fetchBySearch(price *prices.Price) error {
	symbol := strings.ToUpper(price.Asset.Symbol)

	query := strings.ToLower(price.Asset.Symbol)
	if price.Asset.Name != "" && price.Asset.Name != strings.ToLower(price.Asset.Symbol) {
		query = strings.ToLower(price.Asset.Name)
	}

	endpoint := fmt.Sprintf("%s/search/pools?query=%s&network=%s&page=1",
		g.BaseURL, query, g.Network)

	resp, err := g.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				BaseTokenPriceUSD string `json:"base_token_price_usd"`
				Name              string `json:"name"`
				Address           string `json:"address"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("no pools found for %s", symbol)
	}

	priceStr := result.Data[0].Attributes.BaseTokenPriceUSD
	if priceStr == "" {
		return fmt.Errorf("no price available for %s", symbol)
	}

	priceValue, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return fmt.Errorf("invalid price format: %w", err)
	}

	price.Value = priceValue
	price.PoolAddress = result.Data[0].Attributes.Address
	price.Network = g.Network
	return nil
}

func (g *PriceFetcher) FetchMany(pairs ...*prices.Price) error {
	for _, pair := range pairs {
		symbol := strings.ToUpper(pair.Asset.Symbol)
		if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
			pair.Value = 1.0
			continue
		}
		if err := g.Fetch(pair); err != nil {
			continue
		}
	}
	return nil
}

func (g *PriceFetcher) FetchAll() ([]prices.Price, error) {
	endpoint := fmt.Sprintf("%s/networks/%s/trending_pools?page=1", g.BaseURL, g.Network)

	resp, err := g.Client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Attributes struct {
				Name              string `json:"name"`
				BaseTokenPriceUSD string `json:"base_token_price_usd"`
			} `json:"attributes"`
			Relationships struct {
				BaseToken struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"base_token"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	seen := make(map[string]bool)
	pricesList := make([]prices.Price, 0, len(result.Data))

	for _, pool := range result.Data {
		priceStr := pool.Attributes.BaseTokenPriceUSD
		if priceStr == "" {
			continue
		}

		priceValue, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || priceValue == 0 {
			continue
		}

		tokenID := pool.Relationships.BaseToken.Data.ID
		parts := strings.Split(tokenID, "_")
		if len(parts) < 2 {
			continue
		}

		symbol := extractSymbol(pool.Attributes.Name)
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true

		pricesList = append(pricesList, prices.Price{
			Asset: prices.Asset{
				Symbol: strings.ToUpper(symbol),
				Name:   symbol,
			},
			Value: priceValue,
		})
	}

	return pricesList, nil
}

func extractSymbol(poolName string) string {
	parts := strings.Split(poolName, " / ")
	if len(parts) >= 1 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}
