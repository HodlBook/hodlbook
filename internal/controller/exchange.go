package controller

import (
	"net/http"
	"strconv"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

// ListExchanges godoc
// @Summary List exchanges
// @Description Get a list of exchanges with optional filters
// @Tags exchanges
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param asset_id query int false "Asset ID (matches from or to)"
// @Param from_asset_id query int false "From Asset ID"
// @Param to_asset_id query int false "To Asset ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} repo.ExchangeListResult
// @Failure 500 {object} map[string]string
// @Router /api/exchanges [get]
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

// GetExchange godoc
// @Summary Get an exchange by ID
// @Description Get a single exchange by its ID
// @Tags exchanges
// @Produce json
// @Param id path int true "Exchange ID"
// @Success 200 {object} models.Exchange
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/exchanges/{id} [get]
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

// CreateExchange godoc
// @Summary Create a new exchange
// @Description Create a new exchange with the provided data
// @Tags exchanges
// @Accept json
// @Produce json
// @Param exchange body models.Exchange true "Exchange data"
// @Success 201 {object} models.Exchange
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/exchanges [post]
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

// UpdateExchange godoc
// @Summary Update an exchange
// @Description Update an existing exchange by its ID
// @Tags exchanges
// @Accept json
// @Produce json
// @Param id path int true "Exchange ID"
// @Param exchange body models.Exchange true "Exchange data"
// @Success 200 {object} models.Exchange
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/exchanges/{id} [put]
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

// DeleteExchange godoc
// @Summary Delete an exchange
// @Description Delete an exchange by its ID
// @Tags exchanges
// @Param id path int true "Exchange ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/exchanges/{id} [delete]
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
