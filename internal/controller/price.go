package controller

import (
	"net/http"
	"strings"

	"hodlbook/pkg/integrations/prices"

	"github.com/gin-gonic/gin"
)

// ListPrices godoc
// @Summary List current prices
// @Description Get current prices for all tracked assets
// @Tags prices
// @Produce json
// @Success 200 {object} map[string]float64
// @Router /api/prices [get]
func (c *Controller) ListPrices(ctx *gin.Context) {
	if c.priceCache == nil {
		serviceUnavailable(ctx, "price service not available")
		return
	}
	prices := make(map[string]float64)
	for _, key := range c.priceCache.Keys() {
		if val, ok := c.priceCache.Get(key); ok {
			prices[key] = val
		}
	}
	ctx.JSON(http.StatusOK, prices)
}

// GetPrice godoc
// @Summary Get price for a specific asset
// @Description Get current price for a specific asset by symbol
// @Tags prices
// @Produce json
// @Param symbol path string true "Asset symbol (e.g., BTC, ETH)"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/prices/{symbol} [get]
func (c *Controller) GetPrice(ctx *gin.Context) {
	if c.priceCache == nil {
		serviceUnavailable(ctx, "price service not available")
		return
	}
	symbol := ctx.Param("symbol")

	price, ok := c.priceCache.Get(symbol)
	if !ok {
		notFound(ctx, "price not found for symbol")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"price":  price,
	})
}

// GetPriceHistory godoc
// @Summary Get price history for an asset
// @Description Get historical prices for a specific asset by symbol
// @Tags prices
// @Produce json
// @Param symbol path string true "Asset symbol"
// @Success 200 {array} models.AssetHistoricValue
// @Failure 500 {object} map[string]string
// @Router /api/prices/history/{symbol} [get]
func (c *Controller) GetPriceHistory(ctx *gin.Context) {
	symbol := strings.ToUpper(ctx.Param("symbol"))

	history, err := c.repo.SelectAllBySymbol(symbol)
	if err != nil {
		internalError(ctx, "failed to fetch price history")
		return
	}

	ctx.JSON(http.StatusOK, history)
}

// SearchCurrencies godoc
// @Summary Search available currencies from providers
// @Description Get all available currencies from Binance, optionally filtered by search query
// @Tags prices
// @Produce json
// @Param q query string false "Search query to filter currencies"
// @Success 200 {array} object
// @Failure 500 {object} map[string]string
// @Router /api/prices/currencies [get]
func (c *Controller) SearchCurrencies(ctx *gin.Context) {
	query := strings.ToUpper(ctx.Query("q"))

	fetcher := prices.NewPriceService()
	allPrices, err := fetcher.FetchAll()
	if err != nil {
		internalError(ctx, "failed to fetch currencies")
		return
	}

	type CurrencyResult struct {
		Symbol string  `json:"symbol"`
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
	}

	var results []CurrencyResult
	for _, p := range allPrices {
		if query == "" || strings.Contains(p.Asset.Symbol, query) {
			results = append(results, CurrencyResult{
				Symbol: p.Asset.Symbol,
				Name:   p.Asset.Name,
				Price:  p.Value,
			})
		}
	}

	if len(results) > 50 {
		results = results[:50]
	}

	ctx.JSON(http.StatusOK, results)
}

// DeepSearchCurrencies godoc
// @Summary Deep search for currencies across multiple providers
// @Description Search for a currency by symbol/name across DefiLlama and GeckoTerminal
// @Tags prices
// @Produce json
// @Param q query string true "Search query (symbol)"
// @Param name query string false "Asset name for lookup"
// @Param network query string false "Network for direct pool lookup (e.g., base, eth)"
// @Param providers query []string false "Providers to search (defillama, geckoterminal)"
// @Success 200 {array} prices.DeepSearchResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/prices/deep-search [get]
func (c *Controller) DeepSearchCurrencies(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		badRequest(ctx, "query parameter 'q' is required")
		return
	}

	name := ctx.Query("name")
	network := ctx.Query("network")
	providerParams := ctx.QueryArray("providers")

	fetcher := prices.NewPriceService()
	results, err := fetcher.DeepSearch(query, name, network, providerParams)
	if err != nil {
		internalError(ctx, "deep search failed")
		return
	}

	ctx.JSON(http.StatusOK, results)
}

// GetDeepSearchProviders godoc
// @Summary Get available deep search providers
// @Description Get list of available providers for deep search
// @Tags prices
// @Produce json
// @Success 200 {array} string
// @Router /api/prices/deep-search/providers [get]
func (c *Controller) GetDeepSearchProviders(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, prices.AvailableDeepSearchProviders())
}
