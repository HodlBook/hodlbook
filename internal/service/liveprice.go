package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"hodlbook/internal/models"
	priceService "hodlbook/pkg/integrations/prices"
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
	GetUniqueExchangeSymbols() ([]string, error)
}

type assetMeta struct {
	Name        string
	PriceSource string
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
	assetMeta    map[string]assetMeta
	assetMetaMu  sync.RWMutex
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
		assetMeta:    make(map[string]assetMeta),
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

func (s *LivePriceService) ForceSync() error {
	if err := s.syncFromDB(); err != nil {
		return err
	}
	return s.fetchAndPublish()
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

	exchangeSymbols, err := s.repo.GetUniqueExchangeSymbols()
	if err != nil {
		return errors.Wrap(err, "failed to get exchange symbols from DB")
	}

	dbSymbols := make(map[string]bool)
	s.assetMetaMu.Lock()
	for _, asset := range assets {
		dbSymbols[asset.Symbol] = true
		if _, exists := s.cache.Get(asset.Symbol); !exists {
			s.cache.Set(asset.Symbol, 0)
		}
		if asset.PriceSource != nil && *asset.PriceSource != "" {
			s.assetMeta[asset.Symbol] = assetMeta{
				Name:        asset.Name,
				PriceSource: *asset.PriceSource,
			}
		} else if _, exists := s.assetMeta[asset.Symbol]; !exists {
			s.assetMeta[asset.Symbol] = assetMeta{Name: asset.Name}
		}
	}
	s.assetMetaMu.Unlock()

	for _, symbol := range exchangeSymbols {
		dbSymbols[symbol] = true
		if _, exists := s.cache.Get(symbol); !exists {
			s.cache.Set(symbol, 0)
		}
	}

	for _, symbol := range s.cache.Keys() {
		if !dbSymbols[symbol] {
			s.cache.Delete(symbol)
			s.assetMetaMu.Lock()
			delete(s.assetMeta, symbol)
			s.assetMetaMu.Unlock()
		}
	}

	s.lastSync = time.Now()
	s.logger.Info("synced symbols from DB", "assets", len(assets), "exchange_symbols", len(exchangeSymbols))
	return nil
}

func (s *LivePriceService) GetCustomSourceAssets() map[string]assetMeta {
	s.assetMetaMu.RLock()
	defer s.assetMetaMu.RUnlock()

	result := make(map[string]assetMeta)
	for symbol, meta := range s.assetMeta {
		if meta.PriceSource != "" {
			result[symbol] = meta
		}
	}
	return result
}

type CustomSourceAsset struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	PriceSource string  `json:"price_source"`
	Price       float64 `json:"price"`
}

func (s *LivePriceService) GetCustomSourceAssetsWithPrices() []CustomSourceAsset {
	s.assetMetaMu.RLock()
	defer s.assetMetaMu.RUnlock()

	var result []CustomSourceAsset
	for symbol, meta := range s.assetMeta {
		if meta.PriceSource != "" {
			price, _ := s.cache.Get(symbol)
			result = append(result, CustomSourceAsset{
				Symbol:      symbol,
				Name:        meta.Name,
				PriceSource: meta.PriceSource,
				Price:       price,
			})
		}
	}
	return result
}

func (s *LivePriceService) fetchAndPublish() error {
	symbols := s.cache.Keys()
	if len(symbols) == 0 {
		return nil
	}

	s.assetMetaMu.RLock()
	customSourceSymbols := make(map[string]assetMeta)
	regularSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		if meta, ok := s.assetMeta[symbol]; ok && meta.PriceSource != "" {
			customSourceSymbols[symbol] = meta
		} else {
			regularSymbols = append(regularSymbols, symbol)
		}
	}
	s.assetMetaMu.RUnlock()

	priceMap := make(map[string]float64)

	if len(regularSymbols) > 0 {
		pricePairs := make([]*prices.Price, len(regularSymbols))
		for i, symbol := range regularSymbols {
			pricePairs[i] = &prices.Price{
				Asset: prices.Asset{Symbol: symbol},
			}
		}

		if err := s.priceFetcher.FetchMany(pricePairs...); err != nil {
			s.logger.Error("failed to fetch regular prices", "error", err)
		}

		for _, p := range pricePairs {
			s.cache.Set(p.Asset.Symbol, p.Value)
			priceMap[p.Asset.Symbol] = p.Value
		}
	}

	if len(customSourceSymbols) > 0 {
		fetcher := priceService.NewPriceService()
		for symbol, meta := range customSourceSymbols {
			price := &prices.Price{
				Asset: prices.Asset{
					Symbol: symbol,
					Name:   meta.Name,
				},
			}
			if err := fetcher.FetchBySource(meta.PriceSource, price); err != nil {
				s.logger.Debug("failed to fetch custom source price", "symbol", symbol, "source", meta.PriceSource, "error", err)
				continue
			}
			s.cache.Set(symbol, price.Value)
			priceMap[symbol] = price.Value
		}
	}

	data, err := json.Marshal(priceMap)
	if err != nil {
		return errors.Wrap(err, "failed to marshal prices")
	}

	if err := s.publisher.Publish(data); err != nil {
		return errors.Wrap(err, "failed to publish prices")
	}

	s.logger.Debug("published prices", "count", len(priceMap), "custom_sources", len(customSourceSymbols))
	return nil
}

