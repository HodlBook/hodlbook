package service

import (
	"context"
	"log/slog"
	"time"

	"hodlbook/internal/models"
	tickerScheduler "hodlbook/pkg/integrations/scheduler"
	"hodlbook/pkg/types/prices"
	"hodlbook/pkg/types/scheduler"

	"github.com/pkg/errors"
)

var ErrInvalidHistoricPriceConfig = errors.New("invalid historic price service config")

type HistoricValueRepository interface {
	GetUniqueSymbols() ([]string, error)
	Insert(value *models.AssetHistoricValue) error
}

type HistoricPriceService struct {
	ctx          context.Context
	logger       *slog.Logger
	priceFetcher prices.PriceFetcher
	repo         HistoricValueRepository
	scheduler    scheduler.Scheduler
}

type HistoricPriceOption func(*HistoricPriceService)

func WithHistoricPriceContext(ctx context.Context) HistoricPriceOption {
	return func(s *HistoricPriceService) {
		s.ctx = ctx
	}
}

func WithHistoricPriceLogger(l *slog.Logger) HistoricPriceOption {
	return func(s *HistoricPriceService) {
		s.logger = l
	}
}

func WithHistoricPriceFetcher(f prices.PriceFetcher) HistoricPriceOption {
	return func(s *HistoricPriceService) {
		s.priceFetcher = f
	}
}

func WithHistoricPriceRepo(r HistoricValueRepository) HistoricPriceOption {
	return func(s *HistoricPriceService) {
		s.repo = r
	}
}

func (s *HistoricPriceService) IsValid() error {
	switch {
	case s.ctx == nil:
		return errors.Wrap(ErrInvalidHistoricPriceConfig, "ctx cannot be nil")
	case s.logger == nil:
		return errors.Wrap(ErrInvalidHistoricPriceConfig, "logger cannot be nil")
	case s.priceFetcher == nil:
		return errors.Wrap(ErrInvalidHistoricPriceConfig, "price fetcher cannot be nil")
	case s.repo == nil:
		return errors.Wrap(ErrInvalidHistoricPriceConfig, "repo cannot be nil")
	default:
		return nil
	}
}

func NewHistoricPriceService(opts ...HistoricPriceOption) (*HistoricPriceService, error) {
	s := &HistoricPriceService{}

	for _, opt := range opts {
		opt(s)
	}

	if err := s.IsValid(); err != nil {
		return nil, err
	}

	sched, err := tickerScheduler.New(
		tickerScheduler.WithContext(s.ctx),
		tickerScheduler.WithLogger(s.logger),
		tickerScheduler.WithInterval(scheduler.IntervalDaily),
		tickerScheduler.WithHandler(s.tick),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheduler")
	}
	s.scheduler = sched

	return s, nil
}

func (s *HistoricPriceService) Start() error {
	return s.scheduler.Start()
}

func (s *HistoricPriceService) Stop() {
	s.scheduler.Stop()
}

func (s *HistoricPriceService) tick() error {
	symbols, err := s.repo.GetUniqueSymbols()
	if err != nil {
		return errors.Wrap(err, "failed to get symbols from DB")
	}

	if len(symbols) == 0 {
		return nil
	}

	pricePairs := make([]*prices.Price, len(symbols))
	for i, symbol := range symbols {
		pricePairs[i] = &prices.Price{
			Asset: prices.Asset{
				Symbol: symbol,
				Name:   symbol,
			},
		}
	}

	if err := s.priceFetcher.FetchMany(pricePairs...); err != nil {
		return errors.Wrap(err, "failed to fetch prices")
	}

	now := time.Now()
	for i, symbol := range symbols {
		historicValue := &models.AssetHistoricValue{
			Symbol:    symbol,
			Value:     pricePairs[i].Value,
			Timestamp: now,
		}
		if err := s.repo.Insert(historicValue); err != nil {
			s.logger.Error("failed to insert historic value",
				"symbol", symbol,
				"error", err,
			)
			continue
		}
	}

	s.logger.Info("stored historic prices", "count", len(symbols))
	return nil
}
