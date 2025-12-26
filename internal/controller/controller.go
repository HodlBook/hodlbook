package controller

import (
	"hodlbook/pkg/types/cache"
	"hodlbook/pkg/types/repo"
)

type Controller struct {
	repo       repo.Repository
	priceCache cache.Cache[string, float64]
}

type Option func(*Controller)

func WithRepository(r repo.Repository) Option {
	return func(c *Controller) {
		c.repo = r
	}
}

func WithPriceCache(pc cache.Cache[string, float64]) Option {
	return func(c *Controller) {
		c.priceCache = pc
	}
}

func New(opts ...Option) (*Controller, error) {
	c := &Controller{}
	for _, opt := range opts {
		opt(c)
	}
	if c.repo == nil {
		return nil, ErrNilRepository
	}
	return c, nil
}
