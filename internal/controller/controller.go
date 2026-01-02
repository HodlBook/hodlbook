package controller

import (
	"hodlbook/pkg/types/cache"
	"hodlbook/pkg/types/prices"
	"hodlbook/pkg/types/pubsub"
	"hodlbook/pkg/types/repo"
	"log/slog"
)

type Controller struct {
	logger          slog.Logger
	repo            repo.Repository
	priceCache      cache.Cache[string, float64]
	priceFetcher    prices.PriceFetcher
	assetCreatedPub pubsub.Publisher
}

type Option func(*Controller)

func WithLogger(l slog.Logger) Option {
	return func(c *Controller) {
		c.logger = l
	}
}

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

func WithAssetCreatedPublisher(p pubsub.Publisher) Option {
	return func(c *Controller) {
		c.assetCreatedPub = p
	}
}

func WithPriceFetcher(pf prices.PriceFetcher) Option {
	return func(c *Controller) {
		c.priceFetcher = pf
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
