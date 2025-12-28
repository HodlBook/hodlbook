package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"hodlbook/internal/models"
	tickerScheduler "hodlbook/pkg/integrations/scheduler"
	"hodlbook/pkg/types/cache"
	"hodlbook/pkg/types/prices"
	"hodlbook/pkg/types/pubsub"
	"hodlbook/pkg/types/scheduler"

	"github.com/pkg/errors"
)

var ErrInvalidLivePriceConfig = errors.New("invalid live price service config")

type AssetRepository interface {
	GetAllAssets() ([]models.Asset, error)
}

type LivePriceService struct {
	ctx          context.Context
	logger       *slog.Logger
	cache        cache.Cache[string, float64]
	priceFetcher prices.PriceFetcher
	publisher    pubsub.Publisher
	repo         AssetRepository
	scheduler    scheduler.Scheduler
	syncInterval time.Duration
	lastSync     time.Time
}

type LivePriceOption func(*LivePriceService)

func WithLivePriceContext(ctx context.Context) LivePriceOption {
	return func(s *LivePriceService) {
		s.ctx = ctx
	}
}

func WithLivePriceLogger(l *slog.Logger) LivePriceOption {
	return func(s *LivePriceService) {
		s.logger = l
	}
}

func WithLivePriceCache(c cache.Cache[string, float64]) LivePriceOption {
	return func(s *LivePriceService) {
		s.cache = c
	}
}

func WithLivePriceFetcher(f prices.PriceFetcher) LivePriceOption {
	return func(s *LivePriceService) {
		s.priceFetcher = f
	}
}

func WithLivePricePublisher(p pubsub.Publisher) LivePriceOption {
	return func(s *LivePriceService) {
		s.publisher = p
	}
}

func WithLivePriceRepo(r AssetRepository) LivePriceOption {
	return func(s *LivePriceService) {
		s.repo = r
	}
}

func WithLivePriceSyncInterval(d time.Duration) LivePriceOption {
	return func(s *LivePriceService) {
		s.syncInterval = d
	}
}

func (s *LivePriceService) IsValid() error {
	switch {
	case s.ctx == nil:
		return errors.Wrap(ErrInvalidLivePriceConfig, "ctx cannot be nil")
	case s.logger == nil:
		return errors.Wrap(ErrInvalidLivePriceConfig, "logger cannot be nil")
	case s.cache == nil:
		return errors.Wrap(ErrInvalidLivePriceConfig, "cache cannot be nil")
	case s.priceFetcher == nil:
		return errors.Wrap(ErrInvalidLivePriceConfig, "price fetcher cannot be nil")
	case s.publisher == nil:
		return errors.Wrap(ErrInvalidLivePriceConfig, "publisher cannot be nil")
	case s.repo == nil:
		return errors.Wrap(ErrInvalidLivePriceConfig, "repo cannot be nil")
	default:
		return nil
	}
}

func NewLivePriceService(opts ...LivePriceOption) (*LivePriceService, error) {
	s := &LivePriceService{
		syncInterval: time.Hour,
	}

	for _, opt := range opts {
		opt(s)
	}

	if err := s.IsValid(); err != nil {
		return nil, err
	}

	sched, err := tickerScheduler.New(
		tickerScheduler.WithContext(s.ctx),
		tickerScheduler.WithLogger(s.logger),
		tickerScheduler.WithInterval(scheduler.IntervalMinute),
		tickerScheduler.WithHandler(s.tick),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheduler")
	}
	s.scheduler = sched

	return s, nil
}

func (s *LivePriceService) Start() error {
	if err := s.tick(); err != nil {
		s.logger.Error("initial tick failed", "error", err)
	}

	return s.scheduler.Start()
}

func (s *LivePriceService) Stop() {
	s.scheduler.Stop()
}

func (s *LivePriceService) tick() error {
	if len(s.cache.Keys()) == 0 || time.Since(s.lastSync) >= s.syncInterval {
		if err := s.syncFromDB(); err != nil {
			s.logger.Error("DB sync failed", "error", err)
		}
	}

	return s.fetchAndPublish()
}

func (s *LivePriceService) syncFromDB() error {
	assets, err := s.repo.GetAllAssets()
	if err != nil {
		return errors.Wrap(err, "failed to get assets from DB")
	}

	dbSymbols := make(map[string]bool)
	for _, asset := range assets {
		dbSymbols[asset.Symbol] = true
		if _, exists := s.cache.Get(asset.Symbol); !exists {
			s.cache.Set(asset.Symbol, 0)
		}
	}

	for _, symbol := range s.cache.Keys() {
		if !dbSymbols[symbol] {
			s.cache.Delete(symbol)
		}
	}

	s.lastSync = time.Now()
	s.logger.Info("synced assets from DB", "count", len(assets))
	return nil
}

func (s *LivePriceService) fetchAndPublish() error {
	symbols := s.cache.Keys()
	if len(symbols) == 0 {
		return nil
	}

	pricePairs := make([]*prices.Price, len(symbols))
	for i, symbol := range symbols {
		pricePairs[i] = &prices.Price{
			Asset: prices.Asset{Symbol: symbol},
		}
	}

	if err := s.priceFetcher.FetchMany(pricePairs...); err != nil {
		return errors.Wrap(err, "failed to fetch prices")
	}

	priceMap := make(map[string]float64)
	for _, p := range pricePairs {
		s.cache.Set(p.Asset.Symbol, p.Value)
		priceMap[p.Asset.Symbol] = p.Value
	}

	data, err := json.Marshal(priceMap)
	if err != nil {
		return errors.Wrap(err, "failed to marshal prices")
	}

	if err := s.publisher.Publish(data); err != nil {
		return errors.Wrap(err, "failed to publish prices")
	}

	s.logger.Debug("published prices", "count", len(priceMap))
	return nil
}

