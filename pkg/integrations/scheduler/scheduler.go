package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrInvalidSchedulerConfig = errors.New("invalid scheduler config")
)

type Scheduler struct {
	interval time.Duration
	ctx      context.Context
	logger   *slog.Logger
	handler  func() error
	ticker   *time.Ticker
}

type Option func(*Scheduler)

func WithInterval(d time.Duration) Option {
	return func(s *Scheduler) {
		s.interval = d
	}
}

func WithContext(ctx context.Context) Option {
	return func(s *Scheduler) {
		s.ctx = ctx
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(s *Scheduler) {
		s.logger = l
	}
}

func WithHandler(h func() error) Option {
	return func(s *Scheduler) {
		s.handler = h
	}
}

func (s *Scheduler) IsValid() error {
	switch {
	case s.ctx == nil:
		return errors.Wrap(ErrInvalidSchedulerConfig, "ctx cannot be nil")
	case s.logger == nil:
		return errors.Wrap(ErrInvalidSchedulerConfig, "logger cannot be nil")
	case s.interval <= 0:
		return errors.Wrap(ErrInvalidSchedulerConfig, "interval must be positive")
	case s.handler == nil:
		return errors.Wrap(ErrInvalidSchedulerConfig, "handler cannot be nil")
	default:
		return nil
	}
}

func New(opts ...Option) (*Scheduler, error) {
	s := &Scheduler{}

	for _, opt := range opts {
		opt(s)
	}

	return s, s.IsValid()
}

func (s *Scheduler) Start() error {
	if err := s.IsValid(); err != nil {
		return err
	}

	s.ticker = time.NewTicker(s.interval)

	go func() {
		defer s.ticker.Stop()
		for {
			select {
			case <-s.ticker.C:
				if err := s.handler(); err != nil {
					s.logger.Error("scheduler handler error", "interval", s.interval, "error", err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
