package handler

import (
	"net/http"
	"strconv"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

	"github.com/gin-gonic/gin"
)

type ExchangesHandler struct {
	renderer   *Renderer
	repo       *repo.Repository
	priceCache cache.Cache[string, float64]
}

func NewExchangesHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64]) *ExchangesHandler {
	return &ExchangesHandler{
		renderer:   renderer,
		repo:       repository,
		priceCache: priceCache,
	}
}

type ExchangesPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
	Assets     []models.Asset
}

func (h *ExchangesHandler) Index(c *gin.Context) {
	assets, _ := h.repo.GetAllAssets()

	data := ExchangesPageData{
		Title:      "Exchanges",
		PageTitle:  "Exchanges",
		ActivePage: "exchanges",
		Assets:     assets,
	}
	h.renderer.HTML(c, http.StatusOK, "exchanges", data)
}

type ExchangesTableData struct {
	Exchanges  []ExchangeRow
	Empty      bool
	Page       int
	TotalPages int
	HasPrev    bool
	HasNext    bool
}

type ExchangeRow struct {
	ID            int64
	FromAssetID   int64
	FromSymbol    string
	FromAmount    string
	FromAmountRaw float64
	ToAssetID     int64
	ToSymbol      string
	ToAmount      string
	ToAmountRaw   float64
	Fee           string
	FeeRaw        float64
	FeeCurrency   string
	Rate          string
	Timestamp     string
	TimestampRaw  string
	Notes         string
}

func (h *ExchangesHandler) Table(c *gin.Context) {
	fromAssetIDStr := c.Query("from_asset_id")
	toAssetIDStr := c.Query("to_asset_id")
	fromDate := c.Query("from")
	toDate := c.Query("to")
	pageStr := c.DefaultQuery("page", "1")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 20

	var fromAssetID, toAssetID int64
	if fromAssetIDStr != "" {
		fromAssetID, _ = strconv.ParseInt(fromAssetIDStr, 10, 64)
	}
	if toAssetIDStr != "" {
		toAssetID, _ = strconv.ParseInt(toAssetIDStr, 10, 64)
	}

	var fromTime, toTime *time.Time
	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			fromTime = &t
		}
	}
	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
		if err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			toTime = &endOfDay
		}
	}

	filter := repo.ExchangeFilter{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
	if fromAssetID != 0 {
		filter.FromAssetID = &fromAssetID
	}
	if toAssetID != 0 {
		filter.ToAssetID = &toAssetID
	}
	if fromTime != nil {
		filter.StartDate = fromTime
	}
	if toTime != nil {
		filter.EndDate = toTime
	}

	result, err := h.repo.ListExchanges(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load exchanges")
		return
	}
	exchanges := result.Exchanges
	total := int(result.Total)

	assets, _ := h.repo.GetAllAssets()
	assetMap := make(map[int64]string)
	for _, a := range assets {
		assetMap[a.ID] = a.Symbol
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	var rows []ExchangeRow
	for _, ex := range exchanges {
		rate := 0.0
		if ex.FromAmount > 0 {
			rate = ex.ToAmount / ex.FromAmount
		}

		rows = append(rows, ExchangeRow{
			ID:            ex.ID,
			FromAssetID:   ex.FromAssetID,
			FromSymbol:    assetMap[ex.FromAssetID],
			FromAmount:    formatAmount(ex.FromAmount),
			FromAmountRaw: ex.FromAmount,
			ToAssetID:     ex.ToAssetID,
			ToSymbol:      assetMap[ex.ToAssetID],
			ToAmount:      formatAmount(ex.ToAmount),
			ToAmountRaw:   ex.ToAmount,
			Fee:           formatAmount(ex.Fee),
			FeeRaw:        ex.Fee,
			FeeCurrency:   ex.FeeCurrency,
			Rate:          formatAmount(rate),
			Timestamp:     ex.Timestamp.Format("Jan 02, 2006 15:04"),
			TimestampRaw:  ex.Timestamp.Format("2006-01-02T15:04"),
			Notes:         ex.Notes,
		})
	}

	data := ExchangesTableData{
		Exchanges:  rows,
		Empty:      len(rows) == 0,
		Page:       page,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "exchanges_table.html", data)
}

type CreateExchangeRequest struct {
	FromAssetID int64   `form:"from_asset_id" binding:"required"`
	FromAmount  float64 `form:"from_amount" binding:"required"`
	ToAssetID   int64   `form:"to_asset_id" binding:"required"`
	ToAmount    float64 `form:"to_amount" binding:"required"`
	Fee         float64 `form:"fee"`
	FeeCurrency string  `form:"fee_currency"`
	Timestamp   string  `form:"timestamp" binding:"required"`
	Notes       string  `form:"notes"`
}

func (h *ExchangesHandler) Create(c *gin.Context) {
	var req CreateExchangeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid form data", "type": "error"}}`)
		h.Table(c)
		return
	}

	if req.FromAssetID == req.ToAssetID {
		c.Header("HX-Trigger", `{"show-toast": {"message": "From and To assets must be different", "type": "error"}}`)
		h.Table(c)
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.Timestamp)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid date format", "type": "error"}}`)
		h.Table(c)
		return
	}

	exchange := &models.Exchange{
		FromAssetID: req.FromAssetID,
		FromAmount:  req.FromAmount,
		ToAssetID:   req.ToAssetID,
		ToAmount:    req.ToAmount,
		Fee:         req.Fee,
		FeeCurrency: req.FeeCurrency,
		Timestamp:   timestamp,
		Notes:       req.Notes,
	}

	if err := h.repo.CreateExchange(exchange); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to create exchange", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *ExchangesHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid exchange ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	var req CreateExchangeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid form data", "type": "error"}}`)
		h.Table(c)
		return
	}

	if req.FromAssetID == req.ToAssetID {
		c.Header("HX-Trigger", `{"show-toast": {"message": "From and To assets must be different", "type": "error"}}`)
		h.Table(c)
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.Timestamp)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid date format", "type": "error"}}`)
		h.Table(c)
		return
	}

	exchange, err := h.repo.GetExchangeByID(id)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Exchange not found", "type": "error"}}`)
		h.Table(c)
		return
	}

	exchange.FromAssetID = req.FromAssetID
	exchange.FromAmount = req.FromAmount
	exchange.ToAssetID = req.ToAssetID
	exchange.ToAmount = req.ToAmount
	exchange.Fee = req.Fee
	exchange.FeeCurrency = req.FeeCurrency
	exchange.Timestamp = timestamp
	exchange.Notes = req.Notes

	if err := h.repo.UpdateExchange(exchange); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to update exchange", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *ExchangesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid exchange ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	if err := h.repo.DeleteExchange(id); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to delete exchange", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}
