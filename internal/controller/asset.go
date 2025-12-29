package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListAssets godoc
// @Summary List assets
// @Description Get a list of assets with optional filters
// @Tags assets
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param symbol query string false "Symbol"
// @Param transaction_type query string false "Transaction type (deposit, withdraw)"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} repo.AssetListResult
// @Failure 500 {object} map[string]string
// @Router /api/assets [get]
func (c *Controller) ListAssets(ctx *gin.Context) {
	filter := repo.AssetFilter{}

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
	if symbol := ctx.Query("symbol"); symbol != "" {
		filter.Symbol = symbol
	}
	if txType := ctx.Query("transaction_type"); txType != "" {
		filter.TransactionType = txType
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

	result, err := c.repo.ListAssets(filter)
	if err != nil {
		internalError(ctx, "failed to fetch assets")
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// GetAsset godoc
// @Summary Get an asset by ID
// @Description Get a single asset entry by its ID
// @Tags assets
// @Produce json
// @Param id path int true "Asset ID"
// @Success 200 {object} models.Asset
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/assets/{id} [get]
func (c *Controller) GetAsset(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		badRequest(ctx, "invalid asset id")
		return
	}

	asset, err := c.repo.GetAssetByID(id)
	if err != nil {
		notFound(ctx, "asset not found")
		return
	}

	ctx.JSON(http.StatusOK, asset)
}

// CreateAsset godoc
// @Summary Create a new asset entry
// @Description Create a new asset entry with the provided data
// @Tags assets
// @Accept json
// @Produce json
// @Param asset body models.Asset true "Asset data"
// @Success 201 {object} models.Asset
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/assets [post]
func (c *Controller) CreateAsset(ctx *gin.Context) {
	var asset models.Asset
	if err := ctx.ShouldBindJSON(&asset); err != nil {
		badRequestWithDetails(ctx, "invalid input", err.Error())
		return
	}

	if asset.Timestamp.IsZero() {
		asset.Timestamp = time.Now()
	}

	if err := c.repo.CreateAsset(&asset); err != nil {
		internalError(ctx, "failed to create asset")
		return
	}

	ctx.JSON(http.StatusCreated, asset)
}

// UpdateAsset godoc
// @Summary Update an asset entry
// @Description Update an existing asset entry by its ID
// @Tags assets
// @Accept json
// @Produce json
// @Param id path int true "Asset ID"
// @Param asset body models.Asset true "Asset data"
// @Success 200 {object} models.Asset
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/assets/{id} [put]
func (c *Controller) UpdateAsset(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		badRequest(ctx, "invalid asset id")
		return
	}

	if _, err = c.repo.GetAssetByID(id); err != nil {
		notFound(ctx, "asset not found")
		return
	}

	var asset models.Asset
	if err := ctx.ShouldBindJSON(&asset); err != nil {
		badRequestWithDetails(ctx, "invalid input", err.Error())
		return
	}

	asset.ID = id
	if err := c.repo.UpdateAsset(&asset); err != nil {
		internalError(ctx, "failed to update asset")
		return
	}

	ctx.JSON(http.StatusOK, asset)
}

// DeleteAsset godoc
// @Summary Delete an asset entry
// @Description Delete an asset entry by its ID
// @Tags assets
// @Param id path int true "Asset ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/assets/{id} [delete]
func (c *Controller) DeleteAsset(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		badRequest(ctx, "invalid asset id")
		return
	}

	if err := c.repo.DeleteAsset(id); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		internalError(ctx, "failed to delete asset")
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetUniqueSymbols godoc
// @Summary Get unique symbols
// @Description Get a list of unique asset symbols
// @Tags assets
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]string
// @Router /api/assets/symbols [get]
func (c *Controller) GetUniqueSymbols(ctx *gin.Context) {
	symbols, err := c.repo.GetUniqueSymbols()
	if err != nil {
		internalError(ctx, "failed to fetch symbols")
		return
	}
	ctx.JSON(http.StatusOK, symbols)
}
