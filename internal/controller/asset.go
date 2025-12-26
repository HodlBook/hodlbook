package controller

import (
	"net/http"
	"strconv"

	"hodlbook/internal/models"

	"github.com/gin-gonic/gin"
)

// ListAssets godoc
// @Summary List all assets
// @Description Get a list of all assets
// @Tags assets
// @Produce json
// @Success 200 {array} models.Asset
// @Failure 500 {object} map[string]string
// @Router /api/assets [get]
func (c *Controller) ListAssets(ctx *gin.Context) {
	assets, err := c.repo.GetAllAssets()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assets"})
		return
	}
	ctx.JSON(http.StatusOK, assets)
}

// GetAsset godoc
// @Summary Get an asset by ID
// @Description Get a single asset by its ID
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	asset, err := c.repo.GetAssetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	ctx.JSON(http.StatusOK, asset)
}

// CreateAsset godoc
// @Summary Create a new asset
// @Description Create a new asset with the provided data
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if err := c.repo.CreateAsset(&asset); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create asset"})
		return
	}

	ctx.JSON(http.StatusCreated, asset)
}

// DeleteAsset godoc
// @Summary Delete an asset
// @Description Delete an asset by its ID
// @Tags assets
// @Param id path int true "Asset ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /api/assets/{id} [delete]
func (c *Controller) DeleteAsset(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	_ = c.repo.DeleteAsset(id)

	ctx.Status(http.StatusNoContent)
}
