package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) PortfolioSummary(ctx *gin.Context) {
	// TODO: Calculate total portfolio value using transaction and price data
	ctx.JSON(http.StatusOK, gin.H{
		"total_value": 0,
		"currency":    "USD",
	})
}

func (c *Controller) PortfolioAllocation(ctx *gin.Context) {
	// TODO: Calculate portfolio allocation by asset using transaction and price data
	ctx.JSON(http.StatusOK, gin.H{
		"allocations": []gin.H{},
	})
}
