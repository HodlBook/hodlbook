package prices

import (
	"fmt"
	"hodlbook/pkg/common/prices"
	"hodlbook/pkg/integrations/prices/binanceprices"
	"hodlbook/pkg/integrations/prices/coingeckoprices"
)

var (
	_ prices.PriceFetcher = (*PriceService)(nil)
)

type PriceService struct {
	binance   prices.PriceFetcher
	coingecko prices.PriceFetcher
}

func NewPriceService() *PriceService {
	return &PriceService{
		binance:   binanceprices.NewPriceFetcher(),
		coingecko: coingeckoprices.NewPriceFetcher(),
	}
}

func (p *PriceService) Fetch(price *prices.Price) error {
	// Try fetching from Binance first
	err := p.binance.Fetch(price)
	if err == nil {
		return nil
	}
	binanceErr := fmt.Errorf("binance error: %w", err)

	// Fallback to CoinGecko if Binance fails
	err = p.coingecko.Fetch(price)
	if err == nil {
		return nil
	}
	coingeckoErr := fmt.Errorf("coingecko error: %w", err)

	// Merge both errors if both fail
	return fmt.Errorf("%v; %v", binanceErr, coingeckoErr)
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
	return fmt.Errorf("%v; %v", binanceErr, coingeckoErr)
}
