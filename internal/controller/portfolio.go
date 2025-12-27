package controller

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AssetHolding struct {
	AssetID int64   `json:"asset_id"`
	Symbol  string  `json:"symbol"`
	Amount  float64 `json:"amount"`
	Price   float64 `json:"price"`
	Value   float64 `json:"value"`
}

type AllocationEntry struct {
	AssetID    int64   `json:"asset_id"`
	Symbol     string  `json:"symbol"`
	Amount     float64 `json:"amount"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
}

type PerformanceEntry struct {
	AssetID    int64   `json:"asset_id"`
	Symbol     string  `json:"symbol"`
	CostBasis  float64 `json:"cost_basis"`
	CurrentVal float64 `json:"current_value"`
	ProfitLoss float64 `json:"profit_loss"`
	ProfitPct  float64 `json:"profit_percentage"`
}

type HistoryPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

func (c *Controller) calculateHoldings() (map[int64]float64, error) {
	holdings := make(map[int64]float64)

	transactions, err := c.repo.GetAllTransactions()
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

	exchanges, err := c.repo.GetAllExchanges()
	if err != nil {
		return nil, err
	}

	for _, ex := range exchanges {
		holdings[ex.FromAssetID] -= ex.FromAmount
		holdings[ex.ToAssetID] += ex.ToAmount
	}

	return holdings, nil
}

func (c *Controller) getAssetSymbols() (map[int64]string, error) {
	assets, err := c.repo.GetAllAssets()
	if err != nil {
		return nil, err
	}

	symbols := make(map[int64]string)
	for _, asset := range assets {
		symbols[asset.ID] = asset.Symbol
	}
	return symbols, nil
}

// PortfolioSummary godoc
// @Summary Get portfolio summary
// @Description Get the total portfolio value with holdings breakdown
// @Tags portfolio
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/portfolio/summary [get]
func (c *Controller) PortfolioSummary(ctx *gin.Context) {
	holdings, err := c.calculateHoldings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate holdings"})
		return
	}

	symbols, err := c.getAssetSymbols()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assets"})
		return
	}

	var totalValue float64
	assetHoldings := make([]AssetHolding, 0)

	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}

		symbol := symbols[assetID]
		var price float64
		if c.priceCache != nil {
			price, _ = c.priceCache.Get(symbol)
		}

		value := amount * price
		totalValue += value

		assetHoldings = append(assetHoldings, AssetHolding{
			AssetID: assetID,
			Symbol:  symbol,
			Amount:  amount,
			Price:   price,
			Value:   value,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total_value": totalValue,
		"currency":    "USD",
		"holdings":    assetHoldings,
	})
}

// PortfolioAllocation godoc
// @Summary Get portfolio allocation
// @Description Get the portfolio allocation by asset with percentages
// @Tags portfolio
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/portfolio/allocation [get]
func (c *Controller) PortfolioAllocation(ctx *gin.Context) {
	holdings, err := c.calculateHoldings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate holdings"})
		return
	}

	symbols, err := c.getAssetSymbols()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assets"})
		return
	}

	var totalValue float64
	allocations := make([]AllocationEntry, 0)

	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}

		symbol := symbols[assetID]
		var price float64
		if c.priceCache != nil {
			price, _ = c.priceCache.Get(symbol)
		}

		value := amount * price
		totalValue += value

		allocations = append(allocations, AllocationEntry{
			AssetID: assetID,
			Symbol:  symbol,
			Amount:  amount,
			Value:   value,
		})
	}

	for i := range allocations {
		if totalValue > 0 {
			allocations[i].Percentage = (allocations[i].Value / totalValue) * 100
		}
	}

	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].Value > allocations[j].Value
	})

	ctx.JSON(http.StatusOK, gin.H{
		"total_value": totalValue,
		"allocations": allocations,
	})
}

// PortfolioPerformance godoc
// @Summary Get portfolio performance
// @Description Get profit/loss calculations per asset
// @Tags portfolio
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/portfolio/performance [get]
func (c *Controller) PortfolioPerformance(ctx *gin.Context) {
	holdings, err := c.calculateHoldings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate holdings"})
		return
	}

	symbols, err := c.getAssetSymbols()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assets"})
		return
	}

	transactions, err := c.repo.GetAllTransactions()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	assets, err := c.repo.GetAllAssets()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assets"})
		return
	}

	historicPrices := make(map[int64][]struct {
		date  time.Time
		price float64
	})
	for _, asset := range assets {
		history, err := c.repo.SelectAllByAsset(asset.ID)
		if err != nil {
			continue
		}
		for _, h := range history {
			historicPrices[asset.ID] = append(historicPrices[asset.ID], struct {
				date  time.Time
				price float64
			}{h.Timestamp, h.Value})
		}
	}

	findPriceAtTime := func(assetID int64, t time.Time) float64 {
		prices := historicPrices[assetID]
		if len(prices) == 0 {
			return 0
		}
		var closest float64
		minDiff := time.Duration(1<<63 - 1)
		for _, p := range prices {
			diff := t.Sub(p.date)
			if diff < 0 {
				diff = -diff
			}
			if diff < minDiff {
				minDiff = diff
				closest = p.price
			}
		}
		return closest
	}

	exchanges, err := c.repo.GetAllExchanges()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exchanges"})
		return
	}

	type costEvent struct {
		timestamp time.Time
		apply     func(costBasis, runningHoldings map[int64]float64)
	}

	events := make([]costEvent, 0, len(transactions)+len(exchanges))

	for _, tx := range transactions {
		tx := tx
		events = append(events, costEvent{
			timestamp: tx.Timestamp,
			apply: func(costBasis, runningHoldings map[int64]float64) {
				price := findPriceAtTime(tx.AssetID, tx.Timestamp)
				switch tx.Type {
				case "deposit", "buy":
					costBasis[tx.AssetID] += tx.Amount * price
					runningHoldings[tx.AssetID] += tx.Amount
				case "withdraw", "sell":
					if runningHoldings[tx.AssetID] > 0 {
						avgCost := costBasis[tx.AssetID] / runningHoldings[tx.AssetID]
						costBasis[tx.AssetID] -= tx.Amount * avgCost
					}
					runningHoldings[tx.AssetID] -= tx.Amount
				}
			},
		})
	}

	for _, ex := range exchanges {
		ex := ex
		events = append(events, costEvent{
			timestamp: ex.Timestamp,
			apply: func(costBasis, runningHoldings map[int64]float64) {
				var transferredCost float64
				if runningHoldings[ex.FromAssetID] > 0 {
					avgCost := costBasis[ex.FromAssetID] / runningHoldings[ex.FromAssetID]
					transferredCost = ex.FromAmount * avgCost
					costBasis[ex.FromAssetID] -= transferredCost
				}
				runningHoldings[ex.FromAssetID] -= ex.FromAmount

				costBasis[ex.ToAssetID] += transferredCost
				runningHoldings[ex.ToAssetID] += ex.ToAmount
			},
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].timestamp.Before(events[j].timestamp)
	})

	costBasis := make(map[int64]float64)
	runningHoldings := make(map[int64]float64)

	for _, event := range events {
		event.apply(costBasis, runningHoldings)
	}

	var totalCostBasis, totalCurrentValue, totalProfitLoss float64
	performance := make([]PerformanceEntry, 0)

	for assetID, amount := range holdings {
		if amount <= 0 {
			continue
		}

		symbol := symbols[assetID]
		var price float64
		if c.priceCache != nil {
			price, _ = c.priceCache.Get(symbol)
		}

		currentValue := amount * price
		cost := costBasis[assetID]
		profitLoss := currentValue - cost

		var profitPct float64
		if cost > 0 {
			profitPct = (profitLoss / cost) * 100
		}

		totalCostBasis += cost
		totalCurrentValue += currentValue
		totalProfitLoss += profitLoss

		performance = append(performance, PerformanceEntry{
			AssetID:    assetID,
			Symbol:     symbol,
			CostBasis:  cost,
			CurrentVal: currentValue,
			ProfitLoss: profitLoss,
			ProfitPct:  profitPct,
		})
	}

	var totalProfitPct float64
	if totalCostBasis > 0 {
		totalProfitPct = (totalProfitLoss / totalCostBasis) * 100
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total_cost_basis":     totalCostBasis,
		"total_current_value":  totalCurrentValue,
		"total_profit_loss":    totalProfitLoss,
		"total_profit_percent": totalProfitPct,
		"assets":               performance,
	})
}

// PortfolioHistory godoc
// @Summary Get portfolio history
// @Description Get portfolio value over time using historic prices
// @Tags portfolio
// @Produce json
// @Param days query int false "Number of days of history (default 30)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/portfolio/history [get]
func (c *Controller) PortfolioHistory(ctx *gin.Context) {
	days := 30
	if d := ctx.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	assets, err := c.repo.GetAllAssets()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assets"})
		return
	}

	holdings, err := c.calculateHoldings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate holdings"})
		return
	}

	historicPrices := make(map[int64]map[string]float64)
	for _, asset := range assets {
		history, err := c.repo.SelectAllByAsset(asset.ID)
		if err != nil {
			continue
		}

		priceByDate := make(map[string]float64)
		for _, h := range history {
			dateStr := h.Timestamp.Format("2006-01-02")
			priceByDate[dateStr] = h.Value
		}
		historicPrices[asset.ID] = priceByDate
	}

	historyPoints := make([]HistoryPoint, 0)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var dailyValue float64
		for assetID, amount := range holdings {
			if amount <= 0 {
				continue
			}

			if prices, ok := historicPrices[assetID]; ok {
				if price, found := prices[dateStr]; found {
					dailyValue += amount * price
				}
			}
		}

		historyPoints = append(historyPoints, HistoryPoint{
			Date:  dateStr,
			Value: dailyValue,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"days":    days,
		"history": historyPoints,
	})
}
