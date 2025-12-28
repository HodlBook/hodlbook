package handler

import (
	"encoding/json"
	"html/template"
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
	Title        string
	PageTitle    string
	ActivePage   string
	Symbols      []string
	HoldingsJSON template.JS
	PricesJSON   template.JS
}

func (h *ExchangesHandler) Index(c *gin.Context) {
	symbols, _ := h.repo.GetUniqueSymbols()
	holdings := h.calculateHoldings()
	prices := h.getAllPrices()

	holdingsJSON, _ := json.Marshal(holdings)
	pricesJSON, _ := json.Marshal(prices)

	data := ExchangesPageData{
		Title:        "Exchanges",
		PageTitle:    "Exchanges",
		ActivePage:   "exchanges",
		Symbols:      symbols,
		HoldingsJSON: template.JS(holdingsJSON),
		PricesJSON:   template.JS(pricesJSON),
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
	FromSymbol    string
	FromAmount    string
	FromAmountRaw float64
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
	fromSymbol := c.Query("from_symbol")
	toSymbol := c.Query("to_symbol")
	fromDate := c.Query("from")
	toDate := c.Query("to")
	pageStr := c.DefaultQuery("page", "1")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 20

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
	if fromSymbol != "" {
		filter.FromSymbol = &fromSymbol
	}
	if toSymbol != "" {
		filter.ToSymbol = &toSymbol
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

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	var rows []ExchangeRow
	for _, ex := range exchanges {
		var rateStr string
		if ex.FromAmount > 0 && ex.ToAmount > 0 {
			rate := ex.FromAmount / ex.ToAmount
			rateStr = "1 " + ex.ToSymbol + " = " + formatExchangeRate(rate) + " " + ex.FromSymbol
		}

		rows = append(rows, ExchangeRow{
			ID:            ex.ID,
			FromSymbol:    ex.FromSymbol,
			FromAmount:    formatAmount(ex.FromAmount),
			FromAmountRaw: ex.FromAmount,
			ToSymbol:      ex.ToSymbol,
			ToAmount:      formatAmount(ex.ToAmount),
			ToAmountRaw:   ex.ToAmount,
			Fee:           formatAmount(ex.Fee),
			FeeRaw:        ex.Fee,
			FeeCurrency:   ex.FeeCurrency,
			Rate:          rateStr,
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
	FromSymbol  string  `form:"from_symbol" binding:"required"`
	FromAmount  float64 `form:"from_amount" binding:"required"`
	ToSymbol    string  `form:"to_symbol" binding:"required"`
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

	if req.FromSymbol == req.ToSymbol {
		c.Header("HX-Trigger", `{"show-toast": {"message": "From and To symbols must be different", "type": "error"}}`)
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
		FromSymbol:  req.FromSymbol,
		FromAmount:  req.FromAmount,
		ToSymbol:    req.ToSymbol,
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

	if req.FromSymbol == req.ToSymbol {
		c.Header("HX-Trigger", `{"show-toast": {"message": "From and To symbols must be different", "type": "error"}}`)
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

	exchange.FromSymbol = req.FromSymbol
	exchange.FromAmount = req.FromAmount
	exchange.ToSymbol = req.ToSymbol
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

func (h *ExchangesHandler) calculateHoldings() map[string]float64 {
	holdings := make(map[string]float64)

	assets, _ := h.repo.GetAllAssets()

	for _, asset := range assets {
		switch asset.TransactionType {
		case "deposit":
			holdings[asset.Symbol] += asset.Amount
		case "withdraw":
			holdings[asset.Symbol] -= asset.Amount
		}
	}

	exchanges, _ := h.repo.GetAllExchanges()

	for _, ex := range exchanges {
		holdings[ex.FromSymbol] -= ex.FromAmount
		holdings[ex.ToSymbol] += ex.ToAmount
	}

	return holdings
}

func (h *ExchangesHandler) getAllPrices() map[string]float64 {
	prices := make(map[string]float64)
	for _, symbol := range h.priceCache.Keys() {
		price, _ := h.priceCache.Get(symbol)
		prices[symbol] = price
	}
	return prices
}
