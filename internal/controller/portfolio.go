package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PortfolioSummary godoc
// @Summary Get portfolio summary
// @Description Get the total portfolio value
// @Tags portfolio
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/portfolio/summary [get]
func (c *Controller) PortfolioSummary(ctx *gin.Context) {
	// TODO: Calculate total portfolio value using transaction and price data
	ctx.JSON(http.StatusOK, gin.H{
		"total_value": 0,
		"currency":    "USD",
	})
}

// PortfolioAllocation godoc
// @Summary Get portfolio allocation
// @Description Get the portfolio allocation by asset
// @Tags portfolio
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/portfolio/allocation [get]
func (c *Controller) PortfolioAllocation(ctx *gin.Context) {
	// TODO: Calculate portfolio allocation by asset using transaction and price data
	ctx.JSON(http.StatusOK, gin.H{
		"allocations": []gin.H{},
	})
}
