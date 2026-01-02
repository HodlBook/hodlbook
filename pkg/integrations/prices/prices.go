package prices

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"hodlbook/pkg/integrations/prices/binanceprices"
	"hodlbook/pkg/integrations/prices/coingeckoprices"
	"hodlbook/pkg/integrations/prices/cryptocompareprices"
	"hodlbook/pkg/integrations/prices/defillamaprices"
	"hodlbook/pkg/integrations/prices/geckoterminalprices"
	"hodlbook/pkg/integrations/prices/krakenprices"
	"hodlbook/pkg/types/prices"
)

var (
	_ prices.PriceFetcher = (*PriceService)(nil)
)

type cachedPrice struct {
	value     float64
	timestamp time.Time
}

// simple in-memory cache with 1m TTL
type cache struct {
	mu        sync.RWMutex
	timestamp time.Time
	prices    []prices.Price
	bySymbol  map[string]cachedPrice
}

func (c *cache) Get() ([]prices.Price, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Since(c.timestamp) > time.Minute {
		return nil, false
	}
	return c.prices, true
}

func (c *cache) Set(prices []prices.Price) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timestamp = time.Now()
	c.prices = prices
	for _, p := range prices {
		if c.bySymbol == nil {
			c.bySymbol = make(map[string]cachedPrice)
		}
		c.bySymbol[p.Asset.Symbol] = cachedPrice{value: p.Value, timestamp: time.Now()}
	}
}

func (c *cache) GetBySymbol(symbol string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.bySymbol[symbol]
	if !ok || time.Since(p.timestamp) > time.Minute {
		return 0, false
	}
	return p.value, true
}

func (c *cache) SetBySymbol(symbol string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bySymbol == nil {
		c.bySymbol = make(map[string]cachedPrice)
	}
	c.bySymbol[symbol] = cachedPrice{value: value, timestamp: time.Now()}
}

type PriceService struct {
	kraken    prices.PriceFetcher
	binance   prices.PriceFetcher
	coingecko prices.PriceFetcher
	cache     cache
}

func NewPriceService() *PriceService {
	return &PriceService{
		kraken:    krakenprices.NewPriceFetcher(),
		binance:   binanceprices.NewPriceFetcher(),
		coingecko: coingeckoprices.NewPriceFetcher(),
	}
}

func (p *PriceService) Fetch(price *prices.Price) error {
	if cached, ok := p.cache.GetBySymbol(price.Asset.Symbol); ok {
		price.Value = cached
		return nil
	}

	err := p.kraken.Fetch(price)
	if err == nil {
		p.cache.SetBySymbol(price.Asset.Symbol, price.Value)
		return nil
	}
	krakenErr := fmt.Errorf("kraken error: %w", err)

	err = p.binance.Fetch(price)
	if err == nil {
		p.cache.SetBySymbol(price.Asset.Symbol, price.Value)
		return nil
	}
	binanceErr := fmt.Errorf("binance error: %w", err)

	err = p.coingecko.Fetch(price)
	if err == nil {
		p.cache.SetBySymbol(price.Asset.Symbol, price.Value)
		return nil
	}
	coingeckoErr := fmt.Errorf("coingecko error: %w", err)

	return fmt.Errorf("%w; %w; %w", krakenErr, binanceErr, coingeckoErr)
}

func (p *PriceService) FetchMany(pairs ...*prices.Price) error {
	allPrices, err := p.FetchAll()
	if err != nil {
		return err
	}

	priceMap := make(map[string]float64, len(allPrices))
	for _, price := range allPrices {
		priceMap[price.Asset.Symbol] = price.Value
	}

	for _, pair := range pairs {
		if value, ok := priceMap[pair.Asset.Symbol]; ok {
			pair.Value = value
		}
	}

	return nil
}

func (p *PriceService) FetchAll() ([]prices.Price, error) {
	if cached, ok := p.cache.Get(); ok {
		return cached, nil
	}

	merged := make(map[string]prices.Price)

	krakenPrices, krakenErr := p.kraken.FetchAll()
	if krakenErr == nil {
		for _, price := range krakenPrices {
			if price.Value == 0 {
				continue
			}
			symbol := price.Asset.Symbol
			merged[symbol] = price
		}
	}

	binPrices, binErr := p.binance.FetchAll()
	if binErr == nil {
		for _, price := range binPrices {
			if price.Value == 0 {
				continue
			}
			symbol := price.Asset.Symbol
			if _, ok := merged[symbol]; !ok {
				merged[symbol] = price
			}
		}
	}

	cgPrices, cgErr := p.coingecko.FetchAll()
	if cgErr == nil {
		for _, price := range cgPrices {
			if price.Value == 0 {
				continue
			}
			symbol := price.Asset.Symbol
			if existing, ok := merged[symbol]; ok {
				existing.Asset.Name = price.Asset.Name
				merged[symbol] = existing
			} else {
				merged[symbol] = price
			}
		}
	}

	if krakenErr != nil && binErr != nil && cgErr != nil {
		return nil, fmt.Errorf("kraken: %w; binance: %w; coingecko: %w", krakenErr, binErr, cgErr)
	}

	pricesList := make([]prices.Price, 0, len(merged))
	for _, price := range merged {
		pricesList = append(pricesList, price)
	}

	p.cache.Set(pricesList)
	return pricesList, nil
}

var deepSearchProviders = map[string]func() prices.PriceFetcher{
	prices.SourceKraken:        func() prices.PriceFetcher { return krakenprices.NewPriceFetcher() },
	prices.SourceBinance:       func() prices.PriceFetcher { return binanceprices.NewPriceFetcher() },
	prices.SourceCoinGecko:     func() prices.PriceFetcher { return coingeckoprices.NewPriceFetcher() },
	prices.SourceDefiLlama:     func() prices.PriceFetcher { return defillamaprices.NewPriceFetcher() },
	prices.SourceGeckoTerminal: func() prices.PriceFetcher { return geckoterminalprices.NewPriceFetcher() },
}

func AvailableDeepSearchProviders() []string {
	return []string{
		prices.SourceDefiLlama,
		prices.SourceGeckoTerminal,
		prices.SourceKraken,
		prices.SourceBinance,
		prices.SourceCoinGecko,
	}
}

type DeepSearchResult struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Price       float64 `json:"price"`
	Source      string `json:"source"`
	PoolAddress string `json:"pool_address,omitempty"`
	Network     string `json:"network,omitempty"`
}

func (p *PriceService) DeepSearch(query string, name string, network string, providers []string) ([]DeepSearchResult, error) {
	if len(providers) == 0 {
		providers = AvailableDeepSearchProviders()
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = strings.ToLower(query)
	}

	network = strings.TrimSpace(network)

	var results []DeepSearchResult

	for _, providerName := range providers {
		var fetcher prices.PriceFetcher

		if network != "" && isPoolAddress(query) {
			switch providerName {
			case prices.SourceGeckoTerminal:
				fetcher = geckoterminalprices.NewPriceFetcherForNetwork(network)
			case prices.SourceDefiLlama:
				fetcher = defillamaprices.NewPriceFetcherForToken(network, query)
			default:
				factory, ok := deepSearchProviders[providerName]
				if !ok {
					continue
				}
				fetcher = factory()
			}
		} else {
			factory, ok := deepSearchProviders[providerName]
			if !ok {
				continue
			}
			fetcher = factory()
		}

		price := &prices.Price{
			Asset: prices.Asset{
				Symbol: strings.ToUpper(query),
				Name:   name,
			},
		}

		if err := fetcher.Fetch(price); err != nil {
			continue
		}

		if price.Value == 0 {
			continue
		}

		results = append(results, DeepSearchResult{
			Symbol:      strings.ToUpper(price.Asset.Symbol),
			Name:        price.Asset.Name,
			Price:       price.Value,
			Source:      providerName,
			PoolAddress: price.PoolAddress,
			Network:     price.Network,
		})
	}

	return results, nil
}

func isPoolAddress(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "0x") && len(s) == 42
}

func (p *PriceService) FetchBySource(source string, price *prices.Price) error {
	switch source {
	case prices.SourceCryptoCompare:
		return cryptocompareprices.NewPriceFetcher().Fetch(price)
	case prices.SourceDefiLlama:
		return defillamaprices.NewPriceFetcher().Fetch(price)
	case prices.SourceGeckoTerminal:
		return geckoterminalprices.NewPriceFetcher().Fetch(price)
	case prices.SourceCoinGecko:
		return p.coingecko.Fetch(price)
	case prices.SourceBinance:
		return p.binance.Fetch(price)
	case prices.SourceKraken:
		return p.kraken.Fetch(price)
	default:
		return p.Fetch(price)
	}
}
