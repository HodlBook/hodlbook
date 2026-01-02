package krakenprices

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
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		BaseURL: "https://api.kraken.com/0/public",
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type tickerResponse struct {
	Error  []string                     `json:"error"`
	Result map[string]tickerResultEntry `json:"result"`
}

type tickerResultEntry struct {
	Ask   []string `json:"a"` // [price, whole_lot_volume, lot_volume]
	Bid   []string `json:"b"`
	Close []string `json:"c"` // [price, lot_volume] - last trade closed
}

func (k *PriceFetcher) Fetch(price *prices.Price) error {
	symbol := strings.ToUpper(price.Asset.Symbol)
	if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
		price.Value = 1.0
		return nil
	}

	pair := toKrakenPair(symbol)
	endpoint := fmt.Sprintf("%s/Ticker?pair=%s", k.BaseURL, pair)

	resp, err := k.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result tickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Error) > 0 {
		return fmt.Errorf("kraken API error: %s", strings.Join(result.Error, ", "))
	}

	for _, entry := range result.Result {
		if len(entry.Close) > 0 {
			priceValue, err := strconv.ParseFloat(entry.Close[0], 64)
			if err != nil {
				return fmt.Errorf("invalid price format: %w", err)
			}
			price.Value = priceValue
			return nil
		}
	}

	return fmt.Errorf("no price found for %s", symbol)
}

func (k *PriceFetcher) FetchMany(pairs ...*prices.Price) error {
	krakenPairs := make([]string, 0, len(pairs))
	pairMap := make(map[string]*prices.Price)

	for _, p := range pairs {
		symbol := strings.ToUpper(p.Asset.Symbol)
		if symbol == "USD" || symbol == "USDT" || symbol == "USDC" {
			p.Value = 1.0
			continue
		}
		kp := toKrakenPair(symbol)
		krakenPairs = append(krakenPairs, kp)
		pairMap[kp] = p
	}

	if len(krakenPairs) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/Ticker?pair=%s", k.BaseURL, strings.Join(krakenPairs, ","))

	resp, err := k.Client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result tickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Error) > 0 {
		return fmt.Errorf("kraken API error: %s", strings.Join(result.Error, ", "))
	}

	for pairName, entry := range result.Result {
		if len(entry.Close) == 0 {
			continue
		}
		priceValue, err := strconv.ParseFloat(entry.Close[0], 64)
		if err != nil {
			continue
		}
		symbol := fromKrakenPair(pairName)
		for kp, p := range pairMap {
			if fromKrakenPair(kp) == symbol {
				p.Value = priceValue
			}
		}
	}

	return nil
}

func (k *PriceFetcher) FetchAll() ([]prices.Price, error) {
	endpoint := fmt.Sprintf("%s/Ticker", k.BaseURL)

	resp, err := k.Client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result tickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf("kraken API error: %s", strings.Join(result.Error, ", "))
	}

	seen := make(map[string]bool)
	pricesList := make([]prices.Price, 0)

	for pairName, entry := range result.Result {
		if !isUSDPair(pairName) {
			continue
		}
		if len(entry.Close) == 0 {
			continue
		}
		priceValue, err := strconv.ParseFloat(entry.Close[0], 64)
		if err != nil {
			continue
		}
		symbol := fromKrakenPair(pairName)
		if seen[symbol] {
			continue
		}
		seen[symbol] = true
		pricesList = append(pricesList, prices.Price{
			Asset: prices.Asset{Name: symbol, Symbol: symbol},
			Value: priceValue,
		})
	}

	return pricesList, nil
}

func toKrakenPair(symbol string) string {
	symbol = strings.ToUpper(symbol)
	if symbol == "BTC" {
		return "XXBTZUSD"
	}
	if symbol == "ETH" {
		return "XETHZUSD"
	}
	return symbol + "USD"
}

func fromKrakenPair(pair string) string {
	pair = strings.ToUpper(pair)
	if strings.HasPrefix(pair, "XXBT") {
		return "BTC"
	}
	if strings.HasPrefix(pair, "XETH") {
		return "ETH"
	}
	if strings.HasPrefix(pair, "X") && len(pair) > 4 {
		pair = pair[1:]
	}
	if strings.HasSuffix(pair, "ZUSD") {
		return pair[:len(pair)-4]
	}
	if strings.HasSuffix(pair, "USD") {
		return pair[:len(pair)-3]
	}
	return pair
}

func isUSDPair(pair string) bool {
	pair = strings.ToUpper(pair)
	return strings.HasSuffix(pair, "USD") || strings.HasSuffix(pair, "ZUSD")
}
