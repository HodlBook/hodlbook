package coingeckoprices

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
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		BaseURL: "https://api.coingecko.com/api/v3",
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// all asset prices are shown in USD cents
func (c *PriceFetcher) Fetch(price *prices.Price) error {
	return c.FetchMany(price)
}

// all asset prices are shown in USD cents
func (c *PriceFetcher) FetchMany(pairs ...*prices.Price) error {
	ids := make([]string, 0)
	cryptoPairs := make([]*prices.Price, 0)

	for _, pair := range pairs {
		symbol := strings.ToUpper(pair.Asset.Symbol)
		if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
			pair.Value = 1.0
			continue
		}
		ids = append(ids, strings.ToLower(pair.Asset.Name))
		cryptoPairs = append(cryptoPairs, pair)
	}

	if len(ids) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd",
		c.BaseURL,
		strings.Join(ids, ","),
	)

	resp, err := c.Client.Get(endpoint)
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

	for _, pair := range cryptoPairs {
		priceValue, ok := result[strings.ToLower(pair.Asset.Name)]["usd"]
		if !ok {
			continue
		}
		pair.Value = priceValue
	}

	return nil
}

const (
	defaultPages   = 10
	defaultPerPage = 250
)

func (c *PriceFetcher) FetchAll() ([]prices.Price, error) {
	return c.FetchAllPages(defaultPages)
}

func (c *PriceFetcher) FetchAllPages(pages int) ([]prices.Price, error) {
	pricesList := make([]prices.Price, 0, pages*defaultPerPage)

	for page := 1; page <= pages; page++ {
		pagePrices, err := c.fetchPage(page)
		if err != nil {
			if page == 1 {
				return nil, err
			}
			break
		}
		if len(pagePrices) == 0 {
			break
		}
		pricesList = append(pricesList, pagePrices...)
	}

	return pricesList, nil
}

func (c *PriceFetcher) fetchPage(page int) ([]prices.Price, error) {
	endpoint := fmt.Sprintf("%s/coins/markets?vs_currency=usd&per_page=%d&page=%d",
		c.BaseURL, defaultPerPage, page)

	resp, err := c.Client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []struct {
		ID     string   `json:"id"`
		Symbol string   `json:"symbol"`
		Price  *float64 `json:"current_price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	pricesList := make([]prices.Price, 0, len(results))
	for _, result := range results {
		if result.Price == nil || *result.Price == 0 {
			continue
		}
		pricesList = append(pricesList, prices.Price{
			Asset: prices.Asset{Name: result.ID, Symbol: strings.ToUpper(result.Symbol)},
			Value: *result.Price,
		})
	}

	return pricesList, nil
}
