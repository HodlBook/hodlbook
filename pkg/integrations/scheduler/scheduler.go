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
	interval     time.Duration
	ctx          context.Context
	logger       *slog.Logger
	handler      func() error
	ticker       *time.Ticker
	targetHour   int           // hour of day to run (0-23), -1 to disable
	initialDelay time.Duration // delay before first tick, 0 to disable
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

func WithTargetHour(hour int) Option {
	return func(s *Scheduler) {
		s.targetHour = hour
	}
}

func WithInitialDelay(d time.Duration) Option {
	return func(s *Scheduler) {
		s.initialDelay = d
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
	s := &Scheduler{
		targetHour: -1,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, s.IsValid()
}

func (s *Scheduler) Start() error {
	if err := s.IsValid(); err != nil {
		return err
	}

	go func() {
		if s.targetHour >= 0 {
			s.runAtTargetHour()
		} else {
			s.runAtInterval()
		}
	}()

	return nil
}

func (s *Scheduler) runAtTargetHour() {
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), s.targetHour, 0, 0, 0, time.UTC)
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}
		delay := next.Sub(now)

		s.logger.Info("scheduler waiting for target hour", "target", next.Format(time.RFC3339), "delay", delay)

		select {
		case <-time.After(delay):
			s.logger.Info("scheduler firing at target hour")
			if err := s.handler(); err != nil {
				s.logger.Error("scheduler handler error", "error", err)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) runAtInterval() {
	s.ticker = time.NewTicker(s.interval)
	defer s.ticker.Stop()

	if s.initialDelay > 0 {
		s.logger.Info("scheduler started with initial delay", "delay", s.initialDelay, "interval", s.interval)
		select {
		case <-time.After(s.initialDelay):
			s.logger.Info("scheduler initial tick firing")
			if err := s.handler(); err != nil {
				s.logger.Error("scheduler handler error", "interval", s.interval, "error", err)
			}
		case <-s.ctx.Done():
			return
		}
	}

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
}

func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
