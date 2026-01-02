package cryptocompareprices

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
	BaseURL string
	Client  *http.Client
	APIKey  string
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		BaseURL: "https://min-api.cryptocompare.com/data",
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func NewPriceFetcherWithKey(apiKey string) *PriceFetcher {
	f := NewPriceFetcher()
	f.APIKey = apiKey
	return f
}

func (c *PriceFetcher) addAuth(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("authorization", "Apikey "+c.APIKey)
	}
}

func (c *PriceFetcher) Fetch(price *prices.Price) error {
	symbol := strings.ToUpper(price.Asset.Symbol)
	if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
		price.Value = 1.0
		return nil
	}

	endpoint := fmt.Sprintf("%s/price?fsym=%s&tsyms=USD", c.BaseURL, symbol)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuth(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if usd, ok := result["USD"]; ok {
		price.Value = usd
		return nil
	}

	return fmt.Errorf("price not found for %s", symbol)
}

func (c *PriceFetcher) FetchMany(pairs ...*prices.Price) error {
	symbols := make([]string, 0, len(pairs))
	symbolToPair := make(map[string]*prices.Price)

	for _, pair := range pairs {
		symbol := strings.ToUpper(pair.Asset.Symbol)
		if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
			pair.Value = 1.0
			continue
		}
		symbols = append(symbols, symbol)
		symbolToPair[symbol] = pair
	}

	if len(symbols) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/pricemulti?fsyms=%s&tsyms=USD", c.BaseURL, strings.Join(symbols, ","))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuth(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for symbol, priceData := range result {
		if pair, ok := symbolToPair[symbol]; ok {
			if usd, ok := priceData["USD"]; ok {
				pair.Value = usd
			}
		}
	}

	return nil
}

func (c *PriceFetcher) FetchAll() ([]prices.Price, error) {
	endpoint := fmt.Sprintf("%s/top/mktcapfull?limit=100&tsym=USD", c.BaseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuth(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			CoinInfo struct {
				Name     string `json:"Name"`
				FullName string `json:"FullName"`
			} `json:"CoinInfo"`
			Raw struct {
				USD struct {
					Price float64 `json:"PRICE"`
				} `json:"USD"`
			} `json:"RAW"`
		} `json:"Data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	pricesList := make([]prices.Price, 0, len(result.Data))
	for _, coin := range result.Data {
		if coin.Raw.USD.Price == 0 {
			continue
		}
		pricesList = append(pricesList, prices.Price{
			Asset: prices.Asset{
				Symbol: strings.ToUpper(coin.CoinInfo.Name),
				Name:   coin.CoinInfo.FullName,
			},
			Value: coin.Raw.USD.Price,
		})
	}

	return pricesList, nil
}
