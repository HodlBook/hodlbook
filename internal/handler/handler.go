package handler

import (
	"errors"

	"hodlbook/internal/controller"
	"hodlbook/internal/repo"
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
	engine            *gin.Engine
	repository        *repo.Repository
	priceCh           <-chan []byte
	priceCHSet        bool
	priceCache        cache.Cache[string, float64]
	assetCreatedPub   pubsub.Publisher
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
	assets.GET("/:id", ctrl.GetAsset)
	assets.DELETE("/:id", ctrl.DeleteAsset)

	transactions := api.Group("/transactions")
	transactions.GET("", ctrl.ListTransactions)
	transactions.POST("", ctrl.CreateTransaction)
	transactions.GET("/:id", ctrl.GetTransaction)
	transactions.PUT("/:id", ctrl.UpdateTransaction)
	transactions.DELETE("/:id", ctrl.DeleteTransaction)

	exchanges := api.Group("/exchanges")
	exchanges.GET("", ctrl.ListExchanges)
	exchanges.POST("", ctrl.CreateExchange)
	exchanges.GET("/:id", ctrl.GetExchange)
	exchanges.PUT("/:id", ctrl.UpdateExchange)
	exchanges.DELETE("/:id", ctrl.DeleteExchange)

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
	prices.GET("/:symbol", ctrl.GetPrice)
	prices.GET("/history/:id", ctrl.GetPriceHistory)

	return nil
}
