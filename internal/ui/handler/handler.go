package handler

import (
	"errors"
	"html/template"
	"path/filepath"

	"hodlbook/internal/repo"
	"hodlbook/pkg/types/cache"

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
	assets := NewAssetsHandler(h.renderer, h.repo, h.priceCache)
	transactions := NewTransactionsHandler(h.renderer, h.repo, h.priceCache)
	exchanges := NewExchangesHandler(h.renderer, h.repo, h.priceCache)
	prices := NewPricesHandler(h.renderer, h.repo, h.priceCache)

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
	h.engine.POST("/partials/assets/create", assets.Create)
	h.engine.DELETE("/partials/assets/delete/:id", assets.Delete)

	h.engine.GET("/transactions", transactions.Index)
	h.engine.GET("/partials/transactions/table", transactions.Table)
	h.engine.POST("/partials/transactions/create", transactions.Create)
	h.engine.POST("/partials/transactions/update/:id", transactions.Update)
	h.engine.DELETE("/partials/transactions/delete/:id", transactions.Delete)

	h.engine.GET("/exchanges", exchanges.Index)
	h.engine.GET("/partials/exchanges/table", exchanges.Table)
	h.engine.POST("/partials/exchanges/create", exchanges.Create)
	h.engine.POST("/partials/exchanges/update/:id", exchanges.Update)
	h.engine.DELETE("/partials/exchanges/delete/:id", exchanges.Delete)

	h.engine.GET("/prices", prices.Index)
	h.engine.GET("/partials/prices/table", prices.Table)

	h.engine.GET("/api/health", Health)

	return nil
}
