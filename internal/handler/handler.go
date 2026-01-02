package handler

import (
	"errors"
	"net/http"

	"hodlbook/internal/controller"
	"hodlbook/internal/repo"
	"hodlbook/internal/service"
	"hodlbook/pkg/types/cache"
	"hodlbook/pkg/types/pubsub"

	"github.com/gin-gonic/gin"
)

var (
	ErrNilEngine       = errors.New("engine is required")
	ErrNilRepository   = errors.New("repository is required")
	ErrNilPriceChannel = errors.New("price channel is required")
)

type Handler struct {
	engine          *gin.Engine
	repository      *repo.Repository
	priceCh         <-chan []byte
	priceCHSet      bool
	priceCache      cache.Cache[string, float64]
	assetCreatedPub pubsub.Publisher
	livePriceSvc    *service.LivePriceService
}

func (h *Handler) IsValid() error {
	if h.engine == nil {
		return ErrNilEngine
	}
	if h.repository == nil {
		return ErrNilRepository
	}
	if h.priceCHSet && h.priceCh == nil {
		return ErrNilPriceChannel
	}
	return nil
}

type Option func(*Handler)

func WithEngine(engine *gin.Engine) Option {
	return func(h *Handler) {
		h.engine = engine
	}
}

func WithRepository(repository *repo.Repository) Option {
	return func(h *Handler) {
		h.repository = repository
	}
}

func WithPriceChannel(ch <-chan []byte) Option {
	return func(h *Handler) {
		h.priceCh = ch
		h.priceCHSet = true
	}
}

func WithPriceCache(pc cache.Cache[string, float64]) Option {
	return func(h *Handler) {
		h.priceCache = pc
	}
}

func WithAssetCreatedPublisher(p pubsub.Publisher) Option {
	return func(h *Handler) {
		h.assetCreatedPub = p
	}
}

func WithLivePriceService(svc *service.LivePriceService) Option {
	return func(h *Handler) {
		h.livePriceSvc = svc
	}
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}
	if err := h.IsValid(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handler) Setup() error {
	ctrl, err := controller.New(
		controller.WithRepository(h.repository),
		controller.WithPriceCache(h.priceCache),
		controller.WithAssetCreatedPublisher(h.assetCreatedPub),
	)
	if err != nil {
		return err
	}

	api := h.engine.Group("/api")

	assets := api.Group("/assets")
	assets.GET("", ctrl.ListAssets)
	assets.POST("", ctrl.CreateAsset)
	assets.GET("/symbols", ctrl.GetUniqueSymbols)
	assets.GET("/export", ctrl.ExportAssets)
	assets.POST("/import", ctrl.ImportAssets)
	assets.GET("/:id", ctrl.GetAsset)
	assets.PUT("/:id", ctrl.UpdateAsset)
	assets.DELETE("/:id", ctrl.DeleteAsset)

	exchanges := api.Group("/exchanges")
	exchanges.GET("", ctrl.ListExchanges)
	exchanges.POST("", ctrl.CreateExchange)
	exchanges.GET("/export", ctrl.ExportExchanges)
	exchanges.GET("/:id", ctrl.GetExchange)
	exchanges.PUT("/:id", ctrl.UpdateExchange)
	exchanges.DELETE("/:id", ctrl.DeleteExchange)

	imports := api.Group("/imports")
	imports.GET("", ctrl.ListImportLogs)
	imports.GET("/:id", ctrl.GetImportLog)
	imports.POST("/:id/retry", ctrl.RetryImport)
	imports.DELETE("/:id", ctrl.DeleteImportLog)

	portfolio := api.Group("/portfolio")
	portfolio.GET("/summary", ctrl.PortfolioSummary)
	portfolio.GET("/allocation", ctrl.PortfolioAllocation)
	portfolio.GET("/performance", ctrl.PortfolioPerformance)
	portfolio.GET("/history", ctrl.PortfolioHistory)

	prices := api.Group("/prices")
	if h.priceCh != nil {
		prices.GET("/stream", controller.SSEPrices(h.priceCh))
	}
	prices.GET("", ctrl.ListPrices)
	prices.GET("/currencies", ctrl.SearchCurrencies)
	prices.GET("/deep-search", ctrl.DeepSearchCurrencies)
	prices.GET("/deep-search/providers", ctrl.GetDeepSearchProviders)
	if h.livePriceSvc != nil {
		prices.GET("/deep-search/debug", h.debugDeepSearchAssets)
		prices.POST("/sync", h.syncPrices)
	}
	prices.GET("/:symbol", ctrl.GetPrice)
	prices.GET("/history/:symbol", ctrl.GetPriceHistory)

	return nil
}

func (h *Handler) debugDeepSearchAssets(ctx *gin.Context) {
	if h.livePriceSvc == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "live price service not available"})
		return
	}
	assets := h.livePriceSvc.GetCustomSourceAssetsWithPrices()
	ctx.JSON(http.StatusOK, gin.H{
		"count":  len(assets),
		"assets": assets,
	})
}

func (h *Handler) syncPrices(ctx *gin.Context) {
	if h.livePriceSvc == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "live price service not available"})
		return
	}
	if err := h.livePriceSvc.ForceSync(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "prices synced"})
}
