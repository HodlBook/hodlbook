package handler

import (
	"errors"

	"hodlbook/internal/controller"
	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

var (
	ErrNilEngine     = errors.New("engine is required")
	ErrNilRepository = errors.New("repository is required")
)

type Handler struct {
	engine     *gin.Engine
	repository *repo.Repository
}

func New(engine *gin.Engine, repository *repo.Repository) (*Handler, error) {
	if engine == nil {
		return nil, ErrNilEngine
	}
	if repository == nil {
		return nil, ErrNilRepository
	}
	return &Handler{engine: engine, repository: repository}, nil
}

func (h *Handler) Setup() error {
	ctrl, err := controller.New(h.repository)
	if err != nil {
		return err
	}

	api := h.engine.Group("/api")

	assets := api.Group("/assets")
	assets.GET("", ctrl.ListAssets)
	assets.POST("", ctrl.CreateAsset)
	assets.GET("/:id", ctrl.GetAsset)
	assets.PUT("/:id", ctrl.UpdateAsset)
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

	return nil
}
