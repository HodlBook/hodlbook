package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"hodlbook/internal/models"
	"hodlbook/pkg/integrations/wmPubsub"
	"hodlbook/pkg/types/prices"
	"hodlbook/pkg/types/pubsub"

	"github.com/pkg/errors"
)

var ErrInvalidAssetHistoricConfig = errors.New("invalid asset historic service config")

type AssetHistoricRepository interface {
	SelectAllBySymbol(symbol string) ([]models.AssetHistoricValue, error)
	Insert(value *models.AssetHistoricValue) error
}

type AssetHistoricService struct {
	ctx          context.Context
	logger       *slog.Logger
	priceFetcher prices.PriceFetcher
	repo         AssetHistoricRepository
	subscriber   pubsub.Subscriber
	ch           chan []byte
}

type AssetHistoricOption func(*AssetHistoricService)

func WithAssetHistoricContext(ctx context.Context) AssetHistoricOption {
	return func(s *AssetHistoricService) {
		s.ctx = ctx
	}
}

func WithAssetHistoricLogger(l *slog.Logger) AssetHistoricOption {
	return func(s *AssetHistoricService) {
		s.logger = l
	}
}

func WithAssetHistoricFetcher(f prices.PriceFetcher) AssetHistoricOption {
	return func(s *AssetHistoricService) {
		s.priceFetcher = f
	}
}

func WithAssetHistoricRepo(r AssetHistoricRepository) AssetHistoricOption {
	return func(s *AssetHistoricService) {
		s.repo = r
	}
}

func WithAssetHistoricChannel(ch chan []byte) AssetHistoricOption {
	return func(s *AssetHistoricService) {
		s.ch = ch
	}
}

func (s *AssetHistoricService) IsValid() error {
	switch {
	case s.ctx == nil:
		return errors.Wrap(ErrInvalidAssetHistoricConfig, "ctx cannot be nil")
	case s.logger == nil:
		return errors.Wrap(ErrInvalidAssetHistoricConfig, "logger cannot be nil")
	case s.priceFetcher == nil:
		return errors.Wrap(ErrInvalidAssetHistoricConfig, "price fetcher cannot be nil")
	case s.repo == nil:
		return errors.Wrap(ErrInvalidAssetHistoricConfig, "repo cannot be nil")
	case s.ch == nil:
		return errors.Wrap(ErrInvalidAssetHistoricConfig, "channel cannot be nil")
	default:
		return nil
	}
}

func NewAssetHistoricService(opts ...AssetHistoricOption) (*AssetHistoricService, error) {
	s := &AssetHistoricService{}

	for _, opt := range opts {
		opt(s)
	}

	if err := s.IsValid(); err != nil {
		return nil, err
	}

	s.subscriber = wmPubsub.New(
		wmPubsub.WithContext(s.ctx),
		wmPubsub.WithLogger(s.logger),
		wmPubsub.WithTopic("asset-created"),
		wmPubsub.WithChannel(s.ch),
		wmPubsub.WithHandler(s.handleAssetCreated),
	)

	return s, nil
}

func (s *AssetHistoricService) Start() error {
	return s.subscriber.Subscribe()
}

func (s *AssetHistoricService) Publisher() pubsub.Publisher {
	return s.subscriber.(pubsub.Publisher)
}

func (s *AssetHistoricService) handleAssetCreated(data []byte) error {
	var asset models.Asset
	if err := json.Unmarshal(data, &asset); err != nil {
		s.logger.Error("failed to unmarshal asset", "error", err)
		return err
	}

	history, err := s.repo.SelectAllBySymbol(asset.Symbol)
	if err != nil {
		s.logger.Error("failed to check existing history", "symbol", asset.Symbol, "error", err)
		return err
	}

	if len(history) > 0 {
		s.logger.Debug("symbol already has historic data", "symbol", asset.Symbol, "count", len(history))
		return nil
	}

	pricePair := &prices.Price{
		Asset: prices.Asset{
			Symbol: asset.Symbol,
			Name:   asset.Name,
		},
	}

	if err := s.priceFetcher.FetchMany(pricePair); err != nil {
		s.logger.Error("failed to fetch price for new asset", "symbol", asset.Symbol, "error", err)
		return err
	}

	historicValue := &models.AssetHistoricValue{
		Symbol:    asset.Symbol,
		Value:     pricePair.Value,
		Timestamp: time.Now(),
	}

	if err := s.repo.Insert(historicValue); err != nil {
		s.logger.Error("failed to insert initial historic value", "symbol", asset.Symbol, "error", err)
		return err
	}

	s.logger.Info("added initial historic price for new asset", "symbol", asset.Symbol, "price", pricePair.Value)
	return nil
}
