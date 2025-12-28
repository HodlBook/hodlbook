package prices

import (
	"fmt"
	"sync"
	"time"

	"hodlbook/pkg/integrations/prices/binanceprices"
	"hodlbook/pkg/integrations/prices/coingeckoprices"
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
	binance   prices.PriceFetcher
	coingecko prices.PriceFetcher
	cache     cache
}

func NewPriceService() *PriceService {
	return &PriceService{
		binance:   binanceprices.NewPriceFetcher(),
		coingecko: coingeckoprices.NewPriceFetcher(),
	}
}

func (p *PriceService) Fetch(price *prices.Price) error {
	if cached, ok := p.cache.GetBySymbol(price.Asset.Symbol); ok {
		price.Value = cached
		return nil
	}

	// Try fetching from Binance first
	err := p.binance.Fetch(price)
	if err == nil {
		p.cache.SetBySymbol(price.Asset.Symbol, price.Value)
		return nil
	}
	binanceErr := fmt.Errorf("binance error: %w", err)

	// Fallback to CoinGecko if Binance fails
	err = p.coingecko.Fetch(price)
	if err == nil {
		p.cache.SetBySymbol(price.Asset.Symbol, price.Value)
		return nil
	}
	coingeckoErr := fmt.Errorf("coingecko error: %w", err)

	// Merge both errors if both fail
	return fmt.Errorf("%w; %w", binanceErr, coingeckoErr)
}

func (p *PriceService) FetchMany(pairs ...*prices.Price) error {
	// Try fetching from Binance first
	err := p.binance.FetchMany(pairs...)
	if err == nil {
		return nil
	}
	binanceErr := fmt.Errorf("binance error: %w", err)

	// Fallback to CoinGecko if Binance fails
	err = p.coingecko.FetchMany(pairs...)
	if err == nil {
		return nil
	}
	coingeckoErr := fmt.Errorf("coingecko error: %w", err)

	// Merge both errors if both fail
	return fmt.Errorf("%w; %w", binanceErr, coingeckoErr)
}

func (p *PriceService) FetchAll() ([]prices.Price, error) {
	if cached, ok := p.cache.Get(); ok {
		return cached, nil
	}

	// Try Binance first
	pricesList, err := p.binance.FetchAll()
	if err == nil {
		p.cache.Set(pricesList)
		return pricesList, nil
	}
	binanceErr := fmt.Errorf("binance error: %w", err)

	// Fallback to CoinGecko if Binance fails
	pricesList, err = p.coingecko.FetchAll()
	if err == nil {
		p.cache.Set(pricesList)
		return pricesList, nil
	}
	coingeckoErr := fmt.Errorf("coingecko error: %w", err)

	return nil, fmt.Errorf("%w; %w", binanceErr, coingeckoErr)
}
