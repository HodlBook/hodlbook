package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	renderer   *Renderer
	repo       *repo.Repository
	priceCache cache.Cache[string, float64]
}

func NewPortfolioHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64]) *PortfolioHandler {
	return &PortfolioHandler{
		renderer:   renderer,
		repo:       repository,
		priceCache: priceCache,
	}
}

type PortfolioPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
}

func (h *PortfolioHandler) Index(c *gin.Context) {
	data := PortfolioPageData{
		Title:      "Portfolio",
		PageTitle:  "Portfolio",
		ActivePage: "portfolio",
	}
	h.renderer.HTML(c, http.StatusOK, "portfolio", data)
}

type PortfolioSummaryData struct {
	TotalInvested    string
	TotalInvestedRaw float64
	CurrentValue     string
	CurrentValueRaw  float64
	TotalPnL         string
	TotalPnLRaw      float64
	TotalPnLPct      string
	IsPositive       bool
}

func (h *PortfolioHandler) Summary(c *gin.Context) {
	holdings, symbols, currentValue, _ := h.calculatePortfolio()
	transactions, _ := h.repo.GetAllTransactions()
	costBasis := h.calculateCostBasis(transactions)

	var totalInvested float64
	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}
		totalInvested += costBasis[assetID]
	}

	totalPnL := currentValue - totalInvested
	var pnlPct float64
	if totalInvested > 0 {
		pnlPct = (totalPnL / totalInvested) * 100
	}

	_ = symbols

	data := PortfolioSummaryData{
		TotalInvested:    formatCurrency(totalInvested, "USD"),
		TotalInvestedRaw: totalInvested,
		CurrentValue:     formatCurrency(currentValue, "USD"),
		CurrentValueRaw:  currentValue,
		TotalPnL:         formatCurrency(totalPnL, "USD"),
		TotalPnLRaw:      totalPnL,
		TotalPnLPct:      formatPercent(pnlPct),
		IsPositive:       totalPnL >= 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "portfolio_summary.html", data)
}

type PortfolioChartData struct {
	Labels     []string
	Values     []float64
	LabelsJSON string
	ValuesJSON string
}

func (h *PortfolioHandler) Chart(c *gin.Context) {
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

	data := PortfolioChartData{
		Labels:     labels,
		Values:     values,
		LabelsJSON: string(labelsJSON),
		ValuesJSON: string(valuesJSON),
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "portfolio_chart.html", data)
}

type HoldingsTableData struct {
	Holdings   []HoldingRow
	TotalValue string
	Empty      bool
}

type HoldingRow struct {
	AssetID    int64
	Symbol     string
	Amount     string
	AmountRaw  float64
	Price      string
	PriceRaw   float64
	Value      string
	ValueRaw   float64
	Change     string
	ChangeRaw  float64
	Allocation string
	AllocRaw   float64
	Positive   bool
}

func (h *PortfolioHandler) Holdings(c *gin.Context) {
	sortBy := c.DefaultQuery("sort", "value")

	holdings, symbols, totalValue, _ := h.calculatePortfolio()
	transactions, _ := h.repo.GetAllTransactions()
	costBasis := h.calculateCostBasis(transactions)

	var rows []HoldingRow
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

		var allocPct float64
		if totalValue > 0 {
			allocPct = (value / totalValue) * 100
		}

		rows = append(rows, HoldingRow{
			AssetID:    assetID,
			Symbol:     symbol,
			Amount:     formatAmount(amount),
			AmountRaw:  amount,
			Price:      formatPrice(price),
			PriceRaw:   price,
			Value:      formatCurrency(value, "USD"),
			ValueRaw:   value,
			Change:     formatPercent(pnlPct),
			ChangeRaw:  pnlPct,
			Allocation: formatPercentNoSign(allocPct),
			AllocRaw:   allocPct,
			Positive:   pnl >= 0,
		})
	}

	switch sortBy {
	case "value":
		sort.Slice(rows, func(i, j int) bool { return rows[i].ValueRaw > rows[j].ValueRaw })
	case "amount":
		sort.Slice(rows, func(i, j int) bool { return rows[i].AmountRaw > rows[j].AmountRaw })
	case "change":
		sort.Slice(rows, func(i, j int) bool { return rows[i].ChangeRaw > rows[j].ChangeRaw })
	case "allocation":
		sort.Slice(rows, func(i, j int) bool { return rows[i].AllocRaw > rows[j].AllocRaw })
	case "symbol":
		sort.Slice(rows, func(i, j int) bool { return rows[i].Symbol < rows[j].Symbol })
	default:
		sort.Slice(rows, func(i, j int) bool { return rows[i].ValueRaw > rows[j].ValueRaw })
	}

	data := HoldingsTableData{
		Holdings:   rows,
		TotalValue: formatCurrency(totalValue, "USD"),
		Empty:      len(rows) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "portfolio_holdings.html", data)
}

type PerformanceTableData struct {
	Assets         []PerformanceRow
	TotalCostBasis string
	TotalValue     string
	TotalPnL       string
	TotalPnLPct    string
	IsPositive     bool
	Empty          bool
}

type PerformanceRow struct {
	Symbol    string
	CostBasis string
	CostRaw   float64
	Value     string
	ValueRaw  float64
	PnL       string
	PnLRaw    float64
	PnLPct    string
	PnLPctRaw float64
	Positive  bool
}

func (h *PortfolioHandler) Performance(c *gin.Context) {
	holdings, symbols, _, _ := h.calculatePortfolio()
	transactions, _ := h.repo.GetAllTransactions()
	costBasis := h.calculateCostBasis(transactions)

	var rows []PerformanceRow
	var totalCost, totalValue, totalPnL float64

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

		totalCost += cost
		totalValue += value
		totalPnL += pnl

		rows = append(rows, PerformanceRow{
			Symbol:    symbol,
			CostBasis: formatCurrency(cost, "USD"),
			CostRaw:   cost,
			Value:     formatCurrency(value, "USD"),
			ValueRaw:  value,
			PnL:       formatCurrency(pnl, "USD"),
			PnLRaw:    pnl,
			PnLPct:    formatPercent(pnlPct),
			PnLPctRaw: pnlPct,
			Positive:  pnl >= 0,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ValueRaw > rows[j].ValueRaw
	})

	var totalPnLPct float64
	if totalCost > 0 {
		totalPnLPct = (totalPnL / totalCost) * 100
	}

	data := PerformanceTableData{
		Assets:         rows,
		TotalCostBasis: formatCurrency(totalCost, "USD"),
		TotalValue:     formatCurrency(totalValue, "USD"),
		TotalPnL:       formatCurrency(totalPnL, "USD"),
		TotalPnLPct:    formatPercent(totalPnLPct),
		IsPositive:     totalPnL >= 0,
		Empty:          len(rows) == 0,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "portfolio_performance.html", data)
}

func (h *PortfolioHandler) calculateHoldings() (map[int64]float64, error) {
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

func (h *PortfolioHandler) getAssetSymbols() (map[int64]string, error) {
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

func (h *PortfolioHandler) calculatePortfolio() (holdings map[int64]float64, symbols map[int64]string, totalValue float64, err error) {
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

func (h *PortfolioHandler) calculateCostBasis(transactions []models.Transaction) map[int64]float64 {
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

func (h *PortfolioHandler) getSymbolForAsset(assetID int64) (string, error) {
	asset, err := h.repo.GetAssetByID(assetID)
	if err != nil {
		return "", err
	}
	return asset.Symbol, nil
}

func formatPercentNoSign(value float64) string {
	return floatToStr2(value) + "%"
}
