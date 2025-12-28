package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	renderer   *Renderer
	repo       *repo.Repository
	priceCache cache.Cache[string, float64]
}

func NewDashboardHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64]) *DashboardHandler {
	return &DashboardHandler{
		renderer:   renderer,
		repo:       repository,
		priceCache: priceCache,
	}
}

type DashboardData struct {
	Title      string
	PageTitle  string
	ActivePage string
}

func (h *DashboardHandler) Index(c *gin.Context) {
	data := DashboardData{
		Title:      "Dashboard",
		PageTitle:  "Dashboard",
		ActivePage: "dashboard",
	}
	h.renderer.HTML(c, http.StatusOK, "dashboard", data)
}

type SummaryData struct {
	TotalValue    string
	TotalValueRaw float64
	AssetCount    int
	TotalPnL      string
	TotalPnLRaw   float64
	PnLPercent    string
	IsPositive    bool
	BestPerformer string
	BestPnL       string
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	holdings, symbols, totalValue, _ := h.calculatePortfolio()

	var totalCost, totalPnL float64
	var bestSymbol string
	var bestPnLPct float64 = -999999

	transactions, _ := h.repo.GetAllTransactions()
	costBasis := h.calculateCostBasis(transactions)

	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}
		symbol := symbols[assetID]
		price, _ := h.priceCache.Get(symbol)
		currentValue := amount * price
		cost := costBasis[assetID]
		pnl := currentValue - cost
		totalCost += cost
		totalPnL += pnl

		var pnlPct float64
		if cost > 0 {
			pnlPct = (pnl / cost) * 100
		}
		if pnlPct > bestPnLPct && cost > 0 && symbol != "USD" {
			bestPnLPct = pnlPct
			bestSymbol = symbol
		}
	}

	var pnlPct float64
	if totalCost > 0 {
		pnlPct = (totalPnL / totalCost) * 100
	}

	data := SummaryData{
		TotalValue:    formatCurrency(totalValue, "USD"),
		TotalValueRaw: totalValue,
		AssetCount:    len(holdings),
		TotalPnL:      formatCurrency(totalPnL, "USD"),
		TotalPnLRaw:   totalPnL,
		PnLPercent:    formatPercent(pnlPct),
		IsPositive:    totalPnL >= 0,
		BestPerformer: bestSymbol,
		BestPnL:       formatPercent(bestPnLPct),
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "dashboard_summary.html", data)
}

type ChartData struct {
	Labels     []string
	Values     []float64
	LabelsJSON string
	ValuesJSON string
}

func (h *DashboardHandler) Chart(c *gin.Context) {
	rangeParam := c.DefaultQuery("range", "30d")
	days := parseDays(rangeParam)

	assets, _ := h.repo.GetAllAssets()
	holdings, _ := h.calculateHoldings()

	assetSymbols := make(map[int64]string)
	historicPrices := make(map[int64]map[string]float64)
	for _, asset := range assets {
		assetSymbols[asset.ID] = asset.Symbol
		history, err := h.repo.SelectAllByAsset(asset.ID)
		if err != nil {
			continue
		}
		priceByDate := make(map[string]float64)
		for _, hp := range history {
			dateStr := hp.Timestamp.Format("2006-01-02")
			priceByDate[dateStr] = hp.Value
		}
		historicPrices[asset.ID] = priceByDate
	}

	var labels []string
	var values []float64
	now := time.Now()
	todayStr := now.Format("2006-01-02")

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		labelStr := date.Format("Jan 2")

		var dailyValue float64
		for assetID, amount := range holdings {
			if amount <= 0 {
				continue
			}
			var price float64
			if prices, ok := historicPrices[assetID]; ok {
				if p, found := prices[dateStr]; found {
					price = p
				}
			}
			if price == 0 && dateStr == todayStr {
				if symbol, ok := assetSymbols[assetID]; ok {
					price, _ = h.priceCache.Get(symbol)
				}
			}
			dailyValue += amount * price
		}

		labels = append(labels, labelStr)
		values = append(values, dailyValue)
	}

	// Filter out leading zeros to show meaningful data
	startIdx := 0
	for i, v := range values {
		if v > 0 {
			startIdx = i
			break
		}
	}
	if startIdx > 0 && startIdx < len(values) {
		labels = labels[startIdx:]
		values = values[startIdx:]
	}

	labelsJSON, _ := json.Marshal(labels)
	valuesJSON, _ := json.Marshal(values)

	data := ChartData{
		Labels:     labels,
		Values:     values,
		LabelsJSON: string(labelsJSON),
		ValuesJSON: string(valuesJSON),
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "dashboard_chart.html", data)
}

type AllocationData struct {
	Items      []AllocationItem
	TotalValue string
}

type AllocationItem struct {
	Symbol     string
	Value      float64
	Percentage float64
	Color      string
}

var chartColors = []string{
	"#f7931a", "#627eea", "#26a17b", "#2775ca", "#e84142",
	"#8247e5", "#00d395", "#ff007a", "#2b6cb0", "#48bb78",
}

func (h *DashboardHandler) Allocation(c *gin.Context) {
	holdings, symbols, totalValue, _ := h.calculatePortfolio()

	var items []AllocationItem
	i := 0
	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}
		symbol := symbols[assetID]
		price, _ := h.priceCache.Get(symbol)
		value := amount * price

		var pct float64
		if totalValue > 0 {
			pct = (value / totalValue) * 100
		}

		items = append(items, AllocationItem{
			Symbol:     symbol,
			Value:      value,
			Percentage: pct,
			Color:      chartColors[i%len(chartColors)],
		})
		i++
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	data := AllocationData{
		Items:      items,
		TotalValue: formatCurrency(totalValue, "USD"),
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "dashboard_allocation.html", data)
}

type HoldingsData struct {
	Items []HoldingItem
	Empty bool
}

type HoldingItem struct {
	Symbol   string
	Amount   string
	Price    string
	Value    string
	Change   string
	Positive bool
}

func (h *DashboardHandler) Holdings(c *gin.Context) {
	holdings, symbols, _, _ := h.calculatePortfolio()
	transactions, _ := h.repo.GetAllTransactions()
	costBasis := h.calculateCostBasis(transactions)

	var items []HoldingItem
	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}
		symbol := symbols[assetID]
		price, _ := h.priceCache.Get(symbol)
		value := amount * price
		cost := costBasis[assetID]
		pnl := value - cost

		var pnlPct float64
		if cost > 0 {
			pnlPct = (pnl / cost) * 100
		}

		items = append(items, HoldingItem{
			Symbol:   symbol,
			Amount:   formatAmount(amount),
			Price:    formatPrice(price),
			Value:    formatCurrency(value, "USD"),
			Change:   formatPercent(pnlPct),
			Positive: pnl >= 0,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	if len(items) > 5 {
		items = items[:5]
	}

	data := HoldingsData{
		Items: items,
		Empty: len(items) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "dashboard_holdings.html", data)
}

type TransactionsData struct {
	Items []TransactionItem
	Empty bool
}

type TransactionItem struct {
	Type      string
	TypeClass string
	Symbol    string
	Amount    string
	Date      string
}

func (h *DashboardHandler) Transactions(c *gin.Context) {
	transactions, _ := h.repo.GetAllTransactions()
	symbols, _ := h.getAssetSymbols()

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.After(transactions[j].Timestamp)
	})

	var items []TransactionItem
	for i, tx := range transactions {
		if i >= 5 {
			break
		}

		typeClass := "neutral"
		switch tx.Type {
		case "buy", "deposit":
			typeClass = "positive"
		case "sell", "withdraw":
			typeClass = "negative"
		}

		items = append(items, TransactionItem{
			Type:      tx.Type,
			TypeClass: typeClass,
			Symbol:    symbols[tx.AssetID],
			Amount:    formatAmount(tx.Amount),
			Date:      tx.Timestamp.Format("Jan 2, 2006"),
		})
	}

	data := TransactionsData{
		Items: items,
		Empty: len(items) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "dashboard_transactions.html", data)
}

func (h *DashboardHandler) calculateHoldings() (map[int64]float64, error) {
	holdings := make(map[int64]float64)

	transactions, err := h.repo.GetAllTransactions()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	for _, ex := range exchanges {
		holdings[ex.FromAssetID] -= ex.FromAmount
		holdings[ex.ToAssetID] += ex.ToAmount
	}

	return holdings, nil
}

func (h *DashboardHandler) getAssetSymbols() (map[int64]string, error) {
	assets, err := h.repo.GetAllAssets()
	if err != nil {
		return nil, err
	}

	symbols := make(map[int64]string)
	for _, asset := range assets {
		symbols[asset.ID] = asset.Symbol
	}
	return symbols, nil
}

func (h *DashboardHandler) calculatePortfolio() (holdings map[int64]float64, symbols map[int64]string, totalValue float64, err error) {
	holdings, err = h.calculateHoldings()
	if err != nil {
		return
	}

	symbols, err = h.getAssetSymbols()
	if err != nil {
		return
	}

	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}
		symbol := symbols[assetID]
		price, _ := h.priceCache.Get(symbol)
		totalValue += amount * price
	}

	return
}

func (h *DashboardHandler) calculateCostBasis(transactions []models.Transaction) map[int64]float64 {
	costBasis := make(map[int64]float64)
	runningHoldings := make(map[int64]float64)

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.Before(transactions[j].Timestamp)
	})

	for _, tx := range transactions {
		switch tx.Type {
		case "deposit", "buy":
			symbol, _ := h.getSymbolForAsset(tx.AssetID)
			price, _ := h.priceCache.Get(symbol)
			costBasis[tx.AssetID] += tx.Amount * price
			runningHoldings[tx.AssetID] += tx.Amount
		case "withdraw", "sell":
			if runningHoldings[tx.AssetID] > 0 {
				avgCost := costBasis[tx.AssetID] / runningHoldings[tx.AssetID]
				costBasis[tx.AssetID] -= tx.Amount * avgCost
			}
			runningHoldings[tx.AssetID] -= tx.Amount
		}
	}

	return costBasis
}

func (h *DashboardHandler) getSymbolForAsset(assetID int64) (string, error) {
	asset, err := h.repo.GetAssetByID(assetID)
	if err != nil {
		return "", err
	}
	return asset.Symbol, nil
}

func parseDays(rangeStr string) int {
	switch rangeStr {
	case "7d":
		return 7
	case "30d":
		return 30
	case "90d":
		return 90
	case "365d", "1y":
		return 365
	case "all":
		return 3650
	default:
		if days, err := strconv.Atoi(rangeStr); err == nil {
			return days
		}
		return 30
	}
}

func formatAmount(value float64) string {
	if value >= 1000 {
		return floatToStr2(value/1000) + "K"
	}
	if value >= 1 {
		return floatToStr2(value)
	}
	if value >= 0.001 {
		return floatToStrN(value, 4)
	}
	return floatToStrN(value, 8)
}

func floatToStrN(v float64, decimals int) string {
	intPart := int(v)
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	fracPart := int((v - float64(intPart)) * multiplier)
	if fracPart < 0 {
		fracPart = -fracPart
	}

	result := intToStr(intPart) + "."
	fracStr := intToStr(fracPart)
	for len(fracStr) < decimals {
		fracStr = "0" + fracStr
	}
	return result + fracStr
}
