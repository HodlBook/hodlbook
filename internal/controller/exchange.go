package controller

import (
	"net/http"
	"strconv"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListExchanges(ctx *gin.Context) {
	filter := repo.ExchangeFilter{}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}
	if assetIDStr := ctx.Query("asset_id"); assetIDStr != "" {
		if assetID, err := strconv.ParseInt(assetIDStr, 10, 64); err == nil {
			filter.AssetID = &assetID
		}
	}
	if fromAssetIDStr := ctx.Query("from_asset_id"); fromAssetIDStr != "" {
		if fromAssetID, err := strconv.ParseInt(fromAssetIDStr, 10, 64); err == nil {
			filter.FromAssetID = &fromAssetID
		}
	}
	if toAssetIDStr := ctx.Query("to_asset_id"); toAssetIDStr != "" {
		if toAssetID, err := strconv.ParseInt(toAssetIDStr, 10, 64); err == nil {
			filter.ToAssetID = &toAssetID
		}
	}
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = endDate.Add(24*time.Hour - time.Second)
			filter.EndDate = &endDate
		}
	}

	result, err := c.repo.ListExchanges(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exchanges"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *Controller) GetExchange(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange ID"})
		return
	}

	exchange, err := c.repo.GetExchangeByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Exchange not found"})
		return
	}

	ctx.JSON(http.StatusOK, exchange)
}

func (c *Controller) CreateExchange(ctx *gin.Context) {
	var exchange models.Exchange
	if err := ctx.ShouldBindJSON(&exchange); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if exchange.FromAssetID == exchange.ToAssetID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "From and To assets must be different"})
		return
	}

	if exchange.Timestamp.IsZero() {
		exchange.Timestamp = time.Now()
	}

	if err := c.repo.CreateExchange(&exchange); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exchange"})
		return
	}

	ctx.JSON(http.StatusCreated, exchange)
}

func (c *Controller) UpdateExchange(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange ID"})
		return
	}

	if _, err = c.repo.GetExchangeByID(id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Exchange not found"})
		return
	}

	var exchange models.Exchange
	if err := ctx.ShouldBindJSON(&exchange); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if exchange.FromAssetID == exchange.ToAssetID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "From and To assets must be different"})
		return
	}

	exchange.ID = id
	if err := c.repo.UpdateExchange(&exchange); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exchange"})
		return
	}

	ctx.JSON(http.StatusOK, exchange)
}

func (c *Controller) DeleteExchange(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange ID"})
		return
	}

	if _, err = c.repo.GetExchangeByID(id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Exchange not found"})
		return
	}

	if err := c.repo.DeleteExchange(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exchange"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
