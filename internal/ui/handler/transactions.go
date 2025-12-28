package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

	"github.com/gin-gonic/gin"
)

type TransactionsHandler struct {
	renderer   *Renderer
	repo       *repo.Repository
	priceCache cache.Cache[string, float64]
}

func NewTransactionsHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64]) *TransactionsHandler {
	return &TransactionsHandler{
		renderer:   renderer,
		repo:       repository,
		priceCache: priceCache,
	}
}

type TransactionsPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
	Assets     []models.Asset
}

func (h *TransactionsHandler) Index(c *gin.Context) {
	assets, _ := h.repo.GetAllAssets()

	data := TransactionsPageData{
		Title:      "Transactions",
		PageTitle:  "Transactions",
		ActivePage: "transactions",
		Assets:     assets,
	}
	h.renderer.HTML(c, http.StatusOK, "transactions", data)
}

type TransactionsTableData struct {
	Transactions []TransactionRow
	Empty        bool
	Page         int
	TotalPages   int
	HasPrev      bool
	HasNext      bool
}

type TransactionRow struct {
	ID        int64
	Type      string
	TypeClass string
	Symbol    string
	AssetID   int64
	Amount    string
	AmountRaw float64
	Date      string
	Timestamp string
	Notes     string
}

func (h *TransactionsHandler) Table(c *gin.Context) {
	assetIDStr := c.Query("asset_id")
	txType := c.Query("type")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	pageStr := c.DefaultQuery("page", "1")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 20

	transactions, _ := h.repo.GetAllTransactions()
	symbols := h.getAssetSymbols()

	var filtered []models.Transaction
	for _, tx := range transactions {
		if assetIDStr != "" {
			assetID, _ := strconv.ParseInt(assetIDStr, 10, 64)
			if tx.AssetID != assetID {
				continue
			}
		}
		if txType != "" && tx.Type != txType {
			continue
		}
		if fromStr != "" {
			from, err := time.Parse("2006-01-02", fromStr)
			if err == nil && tx.Timestamp.Before(from) {
				continue
			}
		}
		if toStr != "" {
			to, err := time.Parse("2006-01-02", toStr)
			if err == nil && tx.Timestamp.After(to.Add(24*time.Hour)) {
				continue
			}
		}
		filtered = append(filtered, tx)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	total := len(filtered)
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginated := filtered[start:end]

	var rows []TransactionRow
	for _, tx := range paginated {
		typeClass := "neutral"
		switch tx.Type {
		case "buy", "deposit":
			typeClass = "positive"
		case "sell", "withdraw":
			typeClass = "negative"
		}

		rows = append(rows, TransactionRow{
			ID:        tx.ID,
			Type:      tx.Type,
			TypeClass: typeClass,
			Symbol:    symbols[tx.AssetID],
			AssetID:   tx.AssetID,
			Amount:    formatAmount(tx.Amount),
			AmountRaw: tx.Amount,
			Date:      tx.Timestamp.Format("Jan 2, 2006 15:04"),
			Timestamp: tx.Timestamp.Format("2006-01-02T15:04"),
			Notes:     tx.Notes,
		})
	}

	data := TransactionsTableData{
		Transactions: rows,
		Empty:        len(rows) == 0,
		Page:         page,
		TotalPages:   totalPages,
		HasPrev:      page > 1,
		HasNext:      page < totalPages,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "transactions_table.html", data)
}

type CreateTransactionRequest struct {
	Type      string  `form:"type" binding:"required"`
	AssetID   int64   `form:"asset_id" binding:"required"`
	Amount    float64 `form:"amount" binding:"required"`
	Timestamp string  `form:"timestamp" binding:"required"`
	Notes     string  `form:"notes"`
}

func (h *TransactionsHandler) Create(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "All required fields must be filled", "type": "error"}}`)
		h.Table(c)
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	tx := &models.Transaction{
		Type:      strings.ToLower(req.Type),
		AssetID:   req.AssetID,
		Amount:    req.Amount,
		Timestamp: timestamp,
		Notes:     req.Notes,
	}

	if err := h.repo.CreateTransaction(tx); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to create transaction", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *TransactionsHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid transaction ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	var req CreateTransactionRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "All required fields must be filled", "type": "error"}}`)
		h.Table(c)
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	tx := &models.Transaction{
		ID:        id,
		Type:      strings.ToLower(req.Type),
		AssetID:   req.AssetID,
		Amount:    req.Amount,
		Timestamp: timestamp,
		Notes:     req.Notes,
	}

	if err := h.repo.UpdateTransaction(tx); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to update transaction", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *TransactionsHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid transaction ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	if err := h.repo.DeleteTransaction(id); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to delete transaction", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *TransactionsHandler) getAssetSymbols() map[int64]string {
	assets, _ := h.repo.GetAllAssets()
	symbols := make(map[int64]string)
	for _, asset := range assets {
		symbols[asset.ID] = asset.Symbol
	}
	return symbols
}
