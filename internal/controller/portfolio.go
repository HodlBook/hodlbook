package controller

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AssetHolding struct {
	Symbol string  `json:"symbol"`
	Amount float64 `json:"amount"`
	Price  float64 `json:"price"`
	Value  float64 `json:"value"`
}

type AllocationEntry struct {
	Symbol     string  `json:"symbol"`
	Amount     float64 `json:"amount"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
}

type PerformanceEntry struct {
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

func (c *Controller) calculateHoldings() (map[string]float64, error) {
	holdings := make(map[string]float64)

	assets, err := c.repo.GetAllAssets()
	if err != nil {
		return nil, err
	}

	for _, asset := range assets {
		switch asset.TransactionType {
		case "deposit":
			holdings[asset.Symbol] += asset.Amount
		case "withdraw":
			holdings[asset.Symbol] -= asset.Amount
		}
	}

	return holdings, nil
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
		internalError(ctx, "failed to calculate holdings")
		return
	}

	var totalValue float64
	assetHoldings := make([]AssetHolding, 0)

	for symbol, amount := range holdings {
		if amount <= 0 {
			continue
		}

		var price float64
		if c.priceCache != nil {
			price, _ = c.priceCache.Get(symbol)
		}

		value := amount * price
		totalValue += value

		assetHoldings = append(assetHoldings, AssetHolding{
			Symbol: symbol,
			Amount: amount,
			Price:  price,
			Value:  value,
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
		internalError(ctx, "failed to calculate holdings")
		return
	}

	var totalValue float64
	allocations := make([]AllocationEntry, 0)

	for symbol, amount := range holdings {
		if amount <= 0 {
			continue
		}

		var price float64
		if c.priceCache != nil {
			price, _ = c.priceCache.Get(symbol)
		}

		value := amount * price
		totalValue += value

		allocations = append(allocations, AllocationEntry{
			Symbol: symbol,
			Amount: amount,
			Value:  value,
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
		internalError(ctx, "failed to calculate holdings")
		return
	}

	assets, err := c.repo.GetAllAssets()
	if err != nil {
		internalError(ctx, "failed to get assets")
		return
	}

	symbols, err := c.repo.GetUniqueSymbols()
	if err != nil {
		internalError(ctx, "failed to get symbols")
		return
	}

	historicPrices := make(map[string][]struct {
		date  time.Time
		price float64
	})
	for _, symbol := range symbols {
		history, err := c.repo.SelectAllBySymbol(symbol)
		if err != nil {
			continue
		}
		for _, h := range history {
			historicPrices[symbol] = append(historicPrices[symbol], struct {
				date  time.Time
				price float64
			}{h.Timestamp, h.Value})
		}
	}

	findPriceAtTime := func(symbol string, t time.Time) float64 {
		prices := historicPrices[symbol]
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

	type costEvent struct {
		timestamp time.Time
		apply     func(costBasis, runningHoldings map[string]float64)
	}

	events := make([]costEvent, 0, len(assets))

	for _, asset := range assets {
		asset := asset
		events = append(events, costEvent{
			timestamp: asset.Timestamp,
			apply: func(costBasis, runningHoldings map[string]float64) {
				price := findPriceAtTime(asset.Symbol, asset.Timestamp)
				switch asset.TransactionType {
				case "deposit":
					costBasis[asset.Symbol] += asset.Amount * price
					runningHoldings[asset.Symbol] += asset.Amount
				case "withdraw":
					if runningHoldings[asset.Symbol] > 0 {
						avgCost := costBasis[asset.Symbol] / runningHoldings[asset.Symbol]
						costBasis[asset.Symbol] -= asset.Amount * avgCost
					}
					runningHoldings[asset.Symbol] -= asset.Amount
				}
			},
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].timestamp.Before(events[j].timestamp)
	})

	costBasis := make(map[string]float64)
	runningHoldings := make(map[string]float64)

	for _, event := range events {
		event.apply(costBasis, runningHoldings)
	}

	var totalCostBasis, totalCurrentValue, totalProfitLoss float64
	performance := make([]PerformanceEntry, 0)

	for symbol, amount := range holdings {
		if amount <= 0 {
			continue
		}

		var price float64
		if c.priceCache != nil {
			price, _ = c.priceCache.Get(symbol)
		}

		currentValue := amount * price
		cost := costBasis[symbol]
		profitLoss := currentValue - cost

		var profitPct float64
		if cost > 0 {
			profitPct = (profitLoss / cost) * 100
		}

		totalCostBasis += cost
		totalCurrentValue += currentValue
		totalProfitLoss += profitLoss

		performance = append(performance, PerformanceEntry{
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

	symbols, err := c.repo.GetUniqueSymbols()
	if err != nil {
		internalError(ctx, "failed to get symbols")
		return
	}

	holdings, err := c.calculateHoldings()
	if err != nil {
		internalError(ctx, "failed to calculate holdings")
		return
	}

	historicPrices := make(map[string]map[string]float64)
	for _, symbol := range symbols {
		history, err := c.repo.SelectAllBySymbol(symbol)
		if err != nil {
			continue
		}

		priceByDate := make(map[string]float64)
		for _, h := range history {
			dateStr := h.Timestamp.Format("2006-01-02")
			priceByDate[dateStr] = h.Value
		}
		historicPrices[symbol] = priceByDate
	}

	historyPoints := make([]HistoryPoint, 0)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var dailyValue float64
		for symbol, amount := range holdings {
			if amount <= 0 {
				continue
			}

			if prices, ok := historicPrices[symbol]; ok {
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
