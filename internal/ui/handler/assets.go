package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

	"github.com/gin-gonic/gin"
)

type AssetsHandler struct {
	renderer   *Renderer
	repo       *repo.Repository
	priceCache cache.Cache[string, float64]
}

func NewAssetsHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64]) *AssetsHandler {
	return &AssetsHandler{
		renderer:   renderer,
		repo:       repository,
		priceCache: priceCache,
	}
}

type AssetsPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
}

func (h *AssetsHandler) Index(c *gin.Context) {
	data := AssetsPageData{
		Title:      "Assets",
		PageTitle:  "Assets",
		ActivePage: "assets",
	}
	h.renderer.HTML(c, http.StatusOK, "assets", data)
}

type AssetsTableData struct {
	Assets []AssetRow
	Empty  bool
}

type AssetRow struct {
	ID       int64
	Symbol   string
	Name     string
	Price    string
	PriceRaw float64
	Holdings string
	Value    string
	ValueRaw float64
}

func (h *AssetsHandler) Table(c *gin.Context) {
	assets, err := h.repo.GetAllAssets()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load assets")
		return
	}

	holdings := h.calculateHoldings()

	var rows []AssetRow
	for _, asset := range assets {
		price, _ := h.priceCache.Get(asset.Symbol)
		amount := holdings[asset.ID]
		value := amount * price

		rows = append(rows, AssetRow{
			ID:       asset.ID,
			Symbol:   asset.Symbol,
			Name:     asset.Name,
			Price:    formatPrice(price),
			PriceRaw: price,
			Holdings: formatAmount(amount),
			Value:    formatCurrency(value, "USD"),
			ValueRaw: value,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ValueRaw > rows[j].ValueRaw
	})

	data := AssetsTableData{
		Assets: rows,
		Empty:  len(rows) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "assets_table.html", data)
}

type CreateAssetRequest struct {
	Symbol string `form:"symbol" binding:"required"`
	Name   string `form:"name" binding:"required"`
}

func (h *AssetsHandler) Create(c *gin.Context) {
	var req CreateAssetRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Symbol and name are required", "type": "error"}}`)
		h.Table(c)
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	req.Name = strings.TrimSpace(req.Name)

	existing, _ := h.repo.GetAssetBySymbol(req.Symbol)
	if existing != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Asset with this symbol already exists", "type": "error"}}`)
		h.Table(c)
		return
	}

	asset := &models.Asset{
		Symbol: req.Symbol,
		Name:   req.Name,
	}

	if err := h.repo.CreateAsset(asset); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to create asset", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *AssetsHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid asset ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	if err := h.repo.DeleteAsset(id); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to delete asset", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *AssetsHandler) calculateHoldings() map[int64]float64 {
	holdings := make(map[int64]float64)

	transactions, err := h.repo.GetAllTransactions()
	if err != nil {
		return holdings
	}

	for _, tx := range transactions {
		switch tx.Type {
		case "deposit", "buy":
			holdings[tx.AssetID] += tx.Amount
		case "withdraw", "sell":
			holdings[tx.AssetID] -= tx.Amount
		}
	}

	exchanges, err := h.repo.GetAllExchanges()
	if err != nil {
		return holdings
	}

	for _, ex := range exchanges {
		holdings[ex.FromAssetID] -= ex.FromAmount
		holdings[ex.ToAssetID] += ex.ToAmount
	}

	return holdings
}
