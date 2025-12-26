package controller

import (
	"net/http"
	"strconv"

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
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Price service not available"})
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
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Price service not available"})
		return
	}
	symbol := ctx.Param("symbol")

	price, ok := c.priceCache.Get(symbol)
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Price not found for symbol"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"price":  price,
	})
}

// GetPriceHistory godoc
// @Summary Get price history for an asset
// @Description Get historical prices for a specific asset by ID
// @Tags prices
// @Produce json
// @Param id path int true "Asset ID"
// @Success 200 {array} models.AssetHistoricValue
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/prices/history/{id} [get]
func (c *Controller) GetPriceHistory(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	history, err := c.repo.SelectAllByAsset(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price history"})
		return
	}

	ctx.JSON(http.StatusOK, history)
}
