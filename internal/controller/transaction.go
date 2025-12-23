package controller

import (
	"net/http"
	"strconv"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListTransactions(ctx *gin.Context) {
	filter := repo.TransactionFilter{}

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
	if txType := ctx.Query("type"); txType != "" {
		filter.Type = txType
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

	result, err := c.repo.ListTransactions(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *Controller) GetTransaction(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	tx, err := c.repo.GetTransactionByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	ctx.JSON(http.StatusOK, tx)
}

func (c *Controller) CreateTransaction(ctx *gin.Context) {
	var tx models.Transaction
	if err := ctx.ShouldBindJSON(&tx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if tx.Timestamp.IsZero() {
		tx.Timestamp = time.Now()
	}

	if err := c.repo.CreateTransaction(&tx); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	ctx.JSON(http.StatusCreated, tx)
}

func (c *Controller) UpdateTransaction(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if _, err = c.repo.GetTransactionByID(id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	var tx models.Transaction
	if err := ctx.ShouldBindJSON(&tx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	tx.ID = id
	if err := c.repo.UpdateTransaction(&tx); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	ctx.JSON(http.StatusOK, tx)
}

func (c *Controller) DeleteTransaction(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if _, err = c.repo.GetTransactionByID(id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if err := c.repo.DeleteTransaction(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
