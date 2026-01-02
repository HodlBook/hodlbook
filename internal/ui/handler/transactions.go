package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"hodlbook/internal/models"
	"hodlbook/internal/repo"
	pricesIntegration "hodlbook/pkg/integrations/prices"
	"hodlbook/pkg/types/cache"
	"hodlbook/pkg/types/prices"

	"github.com/gin-gonic/gin"
)

type AssetsPageHandler struct {
	renderer     *Renderer
	repo         *repo.Repository
	priceCache   cache.Cache[string, float64]
	priceFetcher prices.PriceFetcher
}

func NewAssetsPageHandler(renderer *Renderer, repository *repo.Repository, priceCache cache.Cache[string, float64], priceFetcher prices.PriceFetcher) *AssetsPageHandler {
	return &AssetsPageHandler{
		renderer:     renderer,
		repo:         repository,
		priceCache:   priceCache,
		priceFetcher: priceFetcher,
	}
}

type AssetsPageData struct {
	Title      string
	PageTitle  string
	ActivePage string
	Symbols    []string
}

func (h *AssetsPageHandler) Index(c *gin.Context) {
	symbols, _ := h.repo.GetUniqueSymbols()

	data := AssetsPageData{
		Title:      "Assets",
		PageTitle:  "Assets",
		ActivePage: "assets",
		Symbols:    symbols,
	}
	h.renderer.HTML(c, http.StatusOK, "assets", data)
}

type AssetsTableData struct {
	Assets     []AssetRow
	Empty      bool
	Page       int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	SortBy     string
	SortDir    string
}

type AssetRow struct {
	ID              int64
	Symbol          string
	Name            string
	TransactionType string
	TypeClass       string
	Amount          string
	AmountRaw       float64
	Date            string
	Timestamp       string
	Notes           string
	USDValue        string
	HasUSDValue     bool
	CurrentValue    string
	HasCurrentValue bool
	GainLossUSD     string
	GainLossPercent string
	GainLossClass   string
	Provider        string
}

func (h *AssetsPageHandler) Table(c *gin.Context) {
	symbol := c.Query("symbol")
	txType := c.Query("type")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	pageStr := c.DefaultQuery("page", "1")
	sortBy := c.DefaultQuery("sort", "date")
	sortDir := c.DefaultQuery("dir", "desc")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 20

	assets, _ := h.repo.GetAllAssets()

	var filtered []models.Asset
	for _, asset := range assets {
		if symbol != "" && asset.Symbol != symbol {
			continue
		}
		if txType != "" && asset.TransactionType != txType {
			continue
		}
		if fromStr != "" {
			from, err := time.Parse("2006-01-02", fromStr)
			if err == nil && asset.Timestamp.Before(from) {
				continue
			}
		}
		if toStr != "" {
			to, err := time.Parse("2006-01-02", toStr)
			if err == nil && asset.Timestamp.After(to.Add(24*time.Hour)) {
				continue
			}
		}
		filtered = append(filtered, asset)
	}

	sortAssets(filtered, sortBy, sortDir)

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

	queries := make([]repo.AssetPriceQuery, len(paginated))
	for i, asset := range paginated {
		queries[i] = repo.AssetPriceQuery{
			Symbol:    asset.Symbol,
			Timestamp: asset.Timestamp,
		}
	}
	historicPrices, _ := h.repo.GetPricesAtTimes(queries, "USD")

	// Fetch current prices from cache (includes deep searched assets)
	currentPrices := make(map[string]float64)
	if h.priceCache != nil {
		for _, symbol := range h.priceCache.Keys() {
			if price, ok := h.priceCache.Get(symbol); ok && price > 0 {
				currentPrices[symbol] = price
			}
		}
	}
	// Fallback to priceFetcher for any missing symbols
	if h.priceFetcher != nil && len(currentPrices) == 0 {
		allPrices, err := h.priceFetcher.FetchAll()
		if err == nil {
			for _, p := range allPrices {
				currentPrices[p.Asset.Symbol] = p.Value
			}
		}
	}

	var rows []AssetRow
	for _, asset := range paginated {
		typeClass := "neutral"
		switch asset.TransactionType {
		case "deposit":
			typeClass = "positive"
		case "withdraw":
			typeClass = "negative"
		}

		var usdValue string
		var historicPrice float64
		hasUSDValue := false
		if priceMap, ok := historicPrices[asset.Symbol]; ok {
			if price, ok := priceMap[asset.Timestamp.Unix()]; ok {
				historicPrice = price
				usdValue = formatPrice(asset.Amount * price)
				hasUSDValue = true
			}
		}

		var currentValue, gainLossUSD, gainLossPercent, gainLossClass string
		hasCurrentValue := false
		if currentPrice, ok := currentPrices[asset.Symbol]; ok {
			currentValue = formatPrice(asset.Amount * currentPrice)
			hasCurrentValue = true

			if hasUSDValue && historicPrice > 0 {
				historicValue := asset.Amount * historicPrice
				currentVal := asset.Amount * currentPrice
				diff := currentVal - historicValue
				pct := (diff / historicValue) * 100

				if diff >= 0 {
					gainLossUSD = "+" + formatPrice(diff)
					gainLossPercent = fmt.Sprintf("+%.2f%%", pct)
					gainLossClass = "positive"
				} else {
					gainLossUSD = formatPrice(diff)
					gainLossPercent = fmt.Sprintf("%.2f%%", pct)
					gainLossClass = "negative"
				}
			}
		}

		provider := "auto"
		if asset.PriceSource != nil && *asset.PriceSource != "" {
			provider = formatProviderName(*asset.PriceSource)
		}

		rows = append(rows, AssetRow{
			ID:              asset.ID,
			Symbol:          asset.Symbol,
			Name:            asset.Name,
			TransactionType: asset.TransactionType,
			TypeClass:       typeClass,
			Amount:          formatAmount(asset.Amount),
			AmountRaw:       asset.Amount,
			Date:            asset.Timestamp.Format("2006-01-02 15:04"),
			Timestamp:       asset.Timestamp.Format("2006-01-02T15:04"),
			Notes:           asset.Notes,
			USDValue:        usdValue,
			HasUSDValue:     hasUSDValue,
			CurrentValue:    currentValue,
			HasCurrentValue: hasCurrentValue,
			GainLossUSD:     gainLossUSD,
			GainLossPercent: gainLossPercent,
			GainLossClass:   gainLossClass,
			Provider:        provider,
		})
	}

	data := AssetsTableData{
		Assets:     rows,
		Empty:      len(rows) == 0,
		Page:       page,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		SortBy:     sortBy,
		SortDir:    sortDir,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "assets_table.html", data)
}

func sortAssets(assets []models.Asset, sortBy, sortDir string) {
	sort.Slice(assets, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "date":
			less = assets[i].Timestamp.Before(assets[j].Timestamp)
		case "type":
			less = assets[i].TransactionType < assets[j].TransactionType
		case "symbol":
			less = assets[i].Symbol < assets[j].Symbol
		case "amount":
			less = assets[i].Amount < assets[j].Amount
		default:
			less = assets[i].Timestamp.Before(assets[j].Timestamp)
		}
		if sortDir == "desc" {
			return !less
		}
		return less
	})
}

type CreateAssetRequest struct {
	Symbol          string  `form:"symbol" binding:"required"`
	Name            string  `form:"name"`
	TransactionType string  `form:"type" binding:"required"`
	Amount          float64 `form:"amount" binding:"required"`
	Timestamp       string  `form:"timestamp" binding:"required"`
	Notes           string  `form:"notes"`
	PriceSource     string  `form:"price_source"`
}

func (h *AssetsPageHandler) Create(c *gin.Context) {
	var req CreateAssetRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "All required fields must be filled", "type": "error"}}`)
		h.Table(c)
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	req.TransactionType = strings.ToLower(strings.TrimSpace(req.TransactionType))
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = req.Symbol
	}

	if req.TransactionType != "deposit" && req.TransactionType != "withdraw" {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid type. Must be deposit or withdraw", "type": "error"}}`)
		h.Table(c)
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	var priceSource *string
	if req.PriceSource != "" {
		priceSource = &req.PriceSource
	}

	asset := &models.Asset{
		Symbol:          req.Symbol,
		Name:            req.Name,
		TransactionType: req.TransactionType,
		Amount:          req.Amount,
		Timestamp:       timestamp,
		Notes:           req.Notes,
		PriceSource:     priceSource,
	}

	if err := h.repo.CreateAsset(asset); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to create asset entry", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.ensurePriceAtTimestamp(asset.Symbol, asset.Name, asset.Timestamp, asset.PriceSource)

	h.Table(c)
}

func (h *AssetsPageHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	asset, err := h.repo.GetAssetByID(id)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Asset not found", "type": "error"}}`)
		h.Table(c)
		return
	}

	var req CreateAssetRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "All required fields must be filled", "type": "error"}}`)
		h.Table(c)
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	req.TransactionType = strings.ToLower(strings.TrimSpace(req.TransactionType))
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = req.Symbol
	}

	if req.TransactionType != "deposit" && req.TransactionType != "withdraw" {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid type. Must be deposit or withdraw", "type": "error"}}`)
		h.Table(c)
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	asset.Symbol = req.Symbol
	asset.Name = req.Name
	asset.TransactionType = req.TransactionType
	asset.Amount = req.Amount
	asset.Timestamp = timestamp
	asset.Notes = req.Notes
	if req.PriceSource != "" {
		asset.PriceSource = &req.PriceSource
	} else {
		asset.PriceSource = nil
	}

	if err := h.repo.UpdateAsset(asset); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to update asset entry", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.ensurePriceAtTimestamp(asset.Symbol, asset.Name, asset.Timestamp, asset.PriceSource)

	h.Table(c)
}

func (h *AssetsPageHandler) ensurePriceAtTimestamp(symbol, name string, timestamp time.Time, priceSource *string) {
	if h.priceFetcher == nil {
		return
	}

	var priceValue float64

	if priceSource != nil && *priceSource != "" {
		price := &prices.Price{
			Asset: prices.Asset{
				Symbol: symbol,
				Name:   name,
			},
		}
		fetcher := pricesIntegration.NewPriceService()
		if err := fetcher.FetchBySource(*priceSource, price); err == nil && price.Value > 0 {
			priceValue = price.Value
		}
	}

	if priceValue == 0 {
		allPrices, err := h.priceFetcher.FetchAll()
		if err != nil {
			return
		}
		for _, p := range allPrices {
			if p.Asset.Symbol == symbol {
				priceValue = p.Value
				break
			}
		}
	}

	if priceValue == 0 {
		return
	}

	h.repo.CreatePrice(&models.Price{
		Symbol:    symbol,
		Currency:  "USD",
		Price:     priceValue,
		Timestamp: timestamp,
	})
}

func (h *AssetsPageHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Invalid ID", "type": "error"}}`)
		h.Table(c)
		return
	}

	if err := h.repo.DeleteAsset(id); err != nil {
		c.Header("HX-Trigger", `{"show-toast": {"message": "Failed to delete asset entry", "type": "error"}}`)
		h.Table(c)
		return
	}

	h.Table(c)
}

func (h *AssetsPageHandler) BulkDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.Header("HX-Trigger", `{"show-toast": {"message": "No assets selected", "type": "error"}}`)
		h.Table(c)
		return
	}

	deleted := 0
	for _, id := range req.IDs {
		if err := h.repo.DeleteAsset(id); err == nil {
			deleted++
		}
	}

	c.Header("HX-Trigger", fmt.Sprintf(`{"show-toast": {"message": "%d assets deleted", "type": "success"}}`, deleted))
	h.Table(c)
}

func (h *AssetsPageHandler) GetSymbols(c *gin.Context) {
	symbols, err := h.repo.GetUniqueSymbols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch symbols"})
		return
	}
	c.JSON(http.StatusOK, symbols)
}

func (h *AssetsPageHandler) GetAssets(c *gin.Context) {
	assets, err := h.repo.GetAllAssets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assets"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

type SupportedCrypto struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}

func (h *AssetsPageHandler) GetSupportedCryptos(c *gin.Context) {
	priceList, err := h.priceFetcher.FetchAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch supported cryptos"})
		return
	}

	cryptos := make([]SupportedCrypto, 0, len(priceList))
	for _, p := range priceList {
		cryptos = append(cryptos, SupportedCrypto{
			Symbol: p.Asset.Symbol,
			Name:   p.Asset.Name,
			Price:  p.Value,
		})
	}

	c.JSON(http.StatusOK, cryptos)
}

func formatProviderName(source string) string {
	names := map[string]string{
		"kraken":        "Kraken",
		"binance":       "Binance",
		"coingecko":     "CoinGecko",
		"defillama":     "DefiLlama",
		"geckoterminal": "GeckoTerminal",
	}
	if name, ok := names[source]; ok {
		return name
	}
	return source
}

type AssetsHoldingsData struct {
	Items    []AssetsHoldingItem
	Empty    bool
	SortBy   string
	SortDir  string
	Endpoint string
	Target   string
}

type AssetsHoldingItem struct {
	Symbol    string
	Amount    string
	Price     string
	Value     string
	ValueRaw  float64
	Change    string
	ChangeRaw float64
	Positive  bool
}

func (h *AssetsPageHandler) Holdings(c *gin.Context) {
	sortBy := c.DefaultQuery("sort", "value")
	sortDir := c.DefaultQuery("dir", "desc")

	items := h.buildHoldingsItems()
	sortAssetsHoldingsItems(items, sortBy, sortDir)

	data := AssetsHoldingsData{
		Items:    items,
		Empty:    len(items) == 0,
		SortBy:   sortBy,
		SortDir:  sortDir,
		Endpoint: "/partials/assets/holdings",
		Target:   "#assets-holdings-container",
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "holdings_table.html", data)
}

func (h *AssetsPageHandler) buildHoldingsItems() []AssetsHoldingItem {
	holdings := h.calculateHoldings()
	assets, _ := h.repo.GetAllAssets()
	costBasis := h.calculateCostBasis(assets)

	var items []AssetsHoldingItem
	for symbol, amount := range holdings {
		if amount <= 0 {
			continue
		}
		price, _ := h.priceCache.Get(symbol)
		value := amount * price
		cost := costBasis[symbol]
		pnl := value - cost

		var pnlPct float64
		if cost > 0 {
			pnlPct = (pnl / cost) * 100
		}

		items = append(items, AssetsHoldingItem{
			Symbol:    symbol,
			Amount:    formatAmount(amount),
			Price:     formatPrice(price),
			Value:     formatCurrency(value, "USD"),
			ValueRaw:  value,
			Change:    formatPercent(pnlPct),
			ChangeRaw: pnlPct,
			Positive:  pnl >= 0,
		})
	}

	return items
}

func (h *AssetsPageHandler) calculateHoldings() map[string]float64 {
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

type assetsCostBasisEvent struct {
	timestamp time.Time
	eventType string
	asset     *models.Asset
	exchange  *models.Exchange
}

func (h *AssetsPageHandler) calculateCostBasis(assets []models.Asset) map[string]float64 {
	costBasis := make(map[string]float64)
	runningHoldings := make(map[string]float64)

	var events []assetsCostBasisEvent
	for i := range assets {
		events = append(events, assetsCostBasisEvent{
			timestamp: assets[i].Timestamp,
			eventType: "asset",
			asset:     &assets[i],
		})
	}

	exchanges, _ := h.repo.GetAllExchanges()
	for i := range exchanges {
		events = append(events, assetsCostBasisEvent{
			timestamp: exchanges[i].Timestamp,
			eventType: "exchange",
			exchange:  &exchanges[i],
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].timestamp.Before(events[j].timestamp)
	})

	for _, event := range events {
		switch event.eventType {
		case "asset":
			asset := event.asset
			switch asset.TransactionType {
			case "deposit":
				price := h.getPriceAtTime(asset.Symbol, asset.Timestamp)
				costBasis[asset.Symbol] += asset.Amount * price
				runningHoldings[asset.Symbol] += asset.Amount
			case "withdraw":
				if runningHoldings[asset.Symbol] > 0 {
					avgCost := costBasis[asset.Symbol] / runningHoldings[asset.Symbol]
					costBasis[asset.Symbol] -= asset.Amount * avgCost
				}
				runningHoldings[asset.Symbol] -= asset.Amount
			}
		case "exchange":
			ex := event.exchange
			if runningHoldings[ex.FromSymbol] > 0 {
				avgCost := costBasis[ex.FromSymbol] / runningHoldings[ex.FromSymbol]
				transferredCost := ex.FromAmount * avgCost
				costBasis[ex.FromSymbol] -= transferredCost
				costBasis[ex.ToSymbol] += transferredCost
			}
			runningHoldings[ex.FromSymbol] -= ex.FromAmount
			runningHoldings[ex.ToSymbol] += ex.ToAmount
		}
	}

	return costBasis
}

func (h *AssetsPageHandler) getPriceAtTime(symbol string, timestamp time.Time) float64 {
	priceRecord, err := h.repo.GetPriceAtTime(symbol, "USD", timestamp)
	if err == nil && priceRecord != nil {
		return priceRecord.Price
	}
	price, _ := h.priceCache.Get(symbol)
	return price
}

func sortAssetsHoldingsItems(items []AssetsHoldingItem, sortBy, sortDir string) {
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "asset":
			less = items[i].Symbol < items[j].Symbol
		case "value":
			less = items[i].ValueRaw < items[j].ValueRaw
		case "change":
			less = items[i].ChangeRaw < items[j].ChangeRaw
		default:
			less = items[i].ValueRaw < items[j].ValueRaw
		}
		if sortDir == "desc" {
			return !less
		}
		return less
	})
}
