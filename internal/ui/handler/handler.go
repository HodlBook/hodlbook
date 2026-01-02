package handler

import (
	"errors"
	"html/template"
	"path/filepath"

	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"
	"hodlbook/pkg/types/prices"

	"github.com/gin-gonic/gin"
)

var (
	ErrNilEngine     = errors.New("engine is required")
	ErrNilRepository = errors.New("repository is required")
)

type WebHandler struct {
	engine       *gin.Engine
	repo         *repo.Repository
	priceCache   cache.Cache[string, float64]
	priceFetcher prices.PriceFetcher
	renderer     *Renderer
	templatesDir string
}

type Option func(*WebHandler)

func WithEngine(engine *gin.Engine) Option {
	return func(h *WebHandler) {
		h.engine = engine
	}
}

func WithRepository(repository *repo.Repository) Option {
	return func(h *WebHandler) {
		h.repo = repository
	}
}

func WithPriceCache(pc cache.Cache[string, float64]) Option {
	return func(h *WebHandler) {
		h.priceCache = pc
	}
}

func WithPriceFetcher(pf prices.PriceFetcher) Option {
	return func(h *WebHandler) {
		h.priceFetcher = pf
	}
}

func WithTemplatesDir(dir string) Option {
	return func(h *WebHandler) {
		h.templatesDir = dir
	}
}

func New(opts ...Option) (*WebHandler, error) {
	h := &WebHandler{
		templatesDir: "./internal/ui/templates",
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.engine == nil {
		return nil, ErrNilEngine
	}
	if h.repo == nil {
		return nil, ErrNilRepository
	}
	h.renderer = NewRenderer(h.templatesDir)
	return h, nil
}

func (h *WebHandler) Setup() error {
	h.engine.Static("/static", "./internal/ui/static")

	h.engine.SetFuncMap(template.FuncMap{
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	})

	partialsPath := filepath.Join(h.templatesDir, "partials", "*.html")
	h.engine.LoadHTMLGlob(partialsPath)

	dashboard := NewDashboardHandler(h.renderer, h.repo, h.priceCache)
	portfolio := NewPortfolioHandler(h.renderer, h.repo, h.priceCache)
	assets := NewAssetsPageHandler(h.renderer, h.repo, h.priceCache, h.priceFetcher)
	exchanges := NewExchangesHandler(h.renderer, h.repo, h.priceCache, h.priceFetcher)
	pricesHandler := NewPricesHandler(h.renderer, h.repo, h.priceCache)
	dataHandler := NewDataHandler(h.renderer, h.repo)

	h.engine.GET("/", dashboard.Index)
	h.engine.GET("/partials/dashboard/summary", dashboard.Summary)
	h.engine.GET("/partials/dashboard/chart", dashboard.Chart)
	h.engine.GET("/partials/dashboard/allocation", dashboard.Allocation)
	h.engine.GET("/partials/dashboard/holdings", dashboard.Holdings)
	h.engine.GET("/partials/dashboard/transactions", dashboard.Transactions)

	h.engine.GET("/portfolio", portfolio.Index)
	h.engine.GET("/partials/portfolio/summary", portfolio.Summary)
	h.engine.GET("/partials/portfolio/chart", portfolio.Chart)
	h.engine.GET("/partials/portfolio/holdings", portfolio.Holdings)
	h.engine.GET("/partials/portfolio/performance", portfolio.Performance)

	h.engine.GET("/assets", assets.Index)
	h.engine.GET("/partials/assets/table", assets.Table)
	h.engine.GET("/partials/assets/holdings", assets.Holdings)
	h.engine.POST("/partials/assets/create", assets.Create)
	h.engine.POST("/partials/assets/update/:id", assets.Update)
	h.engine.DELETE("/partials/assets/delete/:id", assets.Delete)
	h.engine.POST("/partials/assets/bulk-delete", assets.BulkDelete)
	h.engine.GET("/api/ui/assets", assets.GetAssets)
	h.engine.GET("/api/ui/cryptos", assets.GetSupportedCryptos)

	h.engine.GET("/exchanges", exchanges.Index)
	h.engine.GET("/partials/exchanges/table", exchanges.Table)
	h.engine.POST("/partials/exchanges/create", exchanges.Create)
	h.engine.POST("/partials/exchanges/update/:id", exchanges.Update)
	h.engine.DELETE("/partials/exchanges/delete/:id", exchanges.Delete)
	h.engine.POST("/partials/exchanges/bulk-delete", exchanges.BulkDelete)
	h.engine.POST("/partials/exchanges/refresh-prices", exchanges.RefreshPrices)
	h.engine.GET("/api/ui/holdings", exchanges.GetHoldings)

	h.engine.GET("/prices", pricesHandler.Index)
	h.engine.GET("/partials/prices/table", pricesHandler.Table)

	h.engine.GET("/data", dataHandler.Index)
	h.engine.GET("/partials/data/import-history", dataHandler.ImportHistory)

	h.engine.GET("/api/health", Health)

	return nil
}
