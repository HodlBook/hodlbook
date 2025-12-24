package controller

import (
	"net/http"
	"strconv"

	"hodlbook/internal/models"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListAssets(ctx *gin.Context) {
	assets, err := c.repo.GetAllAssets()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assets"})
		return
	}
	ctx.JSON(http.StatusOK, assets)
}

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

func (c *Controller) DeleteAsset(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	_ = c.repo.DeleteAsset(id)

	ctx.Status(http.StatusNoContent)
}
