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
	symbols := h.getAllSymbols()
	holdings := h.calculateHoldings()

	symbolNames := make(map[string]string)
	assets, _ := h.repo.GetAllAssets()
	for _, asset := range assets {
		if _, exists := symbolNames[asset.Symbol]; !exists {
			symbolNames[asset.Symbol] = asset.Name
		}
	}

	var rows []PriceRow
	for _, symbol := range symbols {
		price, _ := h.priceCache.Get(symbol)
		amount := holdings[symbol]
		value := amount * price

		name := symbolNames[symbol]
		if name == "" {
			name = symbol
		}

		rows = append(rows, PriceRow{
			Symbol:   symbol,
			Name:     name,
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

	data := PricesTableData{
		Prices: rows,
		Empty:  len(rows) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "prices_table.html", data)
}

func (h *PricesHandler) calculateHoldings() map[string]float64 {
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

func (h *PricesHandler) getAllSymbols() []string {
	symbolSet := make(map[string]bool)

	assetSymbols, _ := h.repo.GetUniqueSymbols()
	for _, s := range assetSymbols {
		symbolSet[s] = true
	}

	exchangeSymbols, _ := h.repo.GetUniqueExchangeSymbols()
	for _, s := range exchangeSymbols {
		symbolSet[s] = true
	}

	symbols := make([]string, 0, len(symbolSet))
	for s := range symbolSet {
		symbols = append(symbols, s)
	}
	return symbols
}
