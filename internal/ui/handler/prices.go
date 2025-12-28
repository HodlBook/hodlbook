package handler

import (
	"net/http"
	"sort"

	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

	"github.com/gin-gonic/gin"
)

type PricesHandler struct {
	renderer   *Renderer
	repo       *repo.Repository
	priceCache cache.Cache[string, float64]
}

func NewPricesHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64]) *PricesHandler {
	return &PricesHandler{
		renderer:   renderer,
		repo:       repository,
		priceCache: priceCache,
	}
}

type PricesPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
}

func (h *PricesHandler) Index(c *gin.Context) {
	data := PricesPageData{
		Title:      "Prices",
		PageTitle:  "Live Prices",
		ActivePage: "prices",
	}
	h.renderer.HTML(c, http.StatusOK, "prices", data)
}

type PricesTableData struct {
	Prices []PriceRow
	Empty  bool
}

type PriceRow struct {
	Symbol   string
	Name     string
	Price    string
	PriceRaw float64
	Holdings string
	Value    string
	ValueRaw float64
}

func (h *PricesHandler) Table(c *gin.Context) {
	assets, err := h.repo.GetAllAssets()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load assets")
		return
	}

	holdings := h.calculateHoldings()

	var rows []PriceRow
	for _, asset := range assets {
		price, _ := h.priceCache.Get(asset.Symbol)
		amount := holdings[asset.ID]
		value := amount * price

		rows = append(rows, PriceRow{
			Symbol:   asset.Symbol,
			Name:     asset.Name,
			Price:    formatCurrency(price, "USD"),
			PriceRaw: price,
			Holdings: formatAmount(amount),
			Value:    formatCurrency(value, "USD"),
			ValueRaw: value,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ValueRaw > rows[j].ValueRaw
	})

	data := PricesTableData{
		Prices: rows,
		Empty:  len(rows) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "prices_table.html", data)
}

func (h *PricesHandler) calculateHoldings() map[int64]float64 {
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
