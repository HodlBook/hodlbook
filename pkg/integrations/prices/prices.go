package prices

import (
	"fmt"
	"sync"
	"time"

	"hodlbook/pkg/integrations/prices/binanceprices"
	"hodlbook/pkg/integrations/prices/coingeckoprices"
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
			symbol := price.Asset.Symbol
			merged[symbol] = price
		}
	}

	binPrices, binErr := p.binance.FetchAll()
	if binErr == nil {
		for _, price := range binPrices {
			symbol := price.Asset.Symbol
			if _, ok := merged[symbol]; !ok {
				merged[symbol] = price
			}
		}
	}

	cgPrices, cgErr := p.coingecko.FetchAll()
	if cgErr == nil {
		for _, price := range cgPrices {
			symbol := price.Asset.Symbol
			if existing, ok := merged[symbol]; ok {
				if existing.Asset.Name == existing.Asset.Symbol {
					existing.Asset.Name = price.Asset.Name
					merged[symbol] = existing
				}
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
