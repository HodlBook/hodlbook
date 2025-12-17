package controllers

import (
	"net/http"

	"hodlbook/internal/repo"

	"github.com/gin-gonic/gin"
)

type PortfolioController struct {
	repo *repo.Repository
}

// Option is the functional options pattern for PortfolioController
type PortfolioControllerOption func(*PortfolioController) error

// NewPortfolio creates a new portfolio controller with options
func NewPortfolio(opts ...PortfolioControllerOption) (*PortfolioController, error) {
	c := &PortfolioController{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithRepository sets the repository
func WithRepository(repository *repo.Repository) PortfolioControllerOption {
	return func(c *PortfolioController) error {
		if repository == nil {
			return ErrNilRepository
		}
		c.repo = repository
		return nil
	}
}

func (ctrl *PortfolioController) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "portfolio/dashboard.html", gin.H{
		"title": "Dashboard",
	})
}

func (ctrl *PortfolioController) Summary(c *gin.Context) {
	// TODO: Calculate total portfolio value using transaction and price data
	c.JSON(http.StatusOK, gin.H{
		"total_value": 0,
		"currency":    "USD",
	})
}

func (ctrl *PortfolioController) Allocation(c *gin.Context) {
	// TODO: Calculate portfolio allocation by asset using transaction and price data
	c.JSON(http.StatusOK, gin.H{
		"allocations": []gin.H{},
	})
}
