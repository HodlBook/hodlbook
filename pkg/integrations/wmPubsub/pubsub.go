package wmPubsub

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
)

var (
	ErrInvalidPubSubConfig = errors.New("invalid pubsub config")
)

type PubSub struct {
	topic   string
	ch      chan []byte
	ctx     context.Context
	logger  *slog.Logger
	handler func([]byte) error
}

type Option func(*PubSub)

func WithContext(ctx context.Context) Option {
	return func(ps *PubSub) {
		ps.ctx = ctx
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(ps *PubSub) {
		ps.logger = l
	}
}

func WithTopic(topic string) Option {
	return func(ps *PubSub) {
		ps.topic = topic
	}
}

func WithHandler(h func([]byte) error) Option {
	return func(ps *PubSub) {
		ps.handler = h
	}
}

func WithChannel(ch chan []byte) Option {
	return func(ps *PubSub) {
		ps.ch = ch
	}
}

func (ps *PubSub) IsValid() error {
	switch {
	case ps.ctx == nil:
		return errors.Wrap(ErrInvalidPubSubConfig, "ctx cannot be nil")
	case ps.logger == nil:
		return errors.Wrap(ErrInvalidPubSubConfig, "logger cannot be nil")
	case ps.topic == "":
		return errors.Wrap(ErrInvalidPubSubConfig, "topic cannot be empty")
	case ps.ch == nil:
		return errors.Wrap(ErrInvalidPubSubConfig, "channel cannot be nil")
	default:
		return nil
	}
}

func New(opts ...Option) *PubSub {
	ps := &PubSub{}

	for _, opt := range opts {
		opt(ps)
	}

	return ps
}

func (ps *PubSub) Publish(payload []byte) error {
	select {
	case ps.ch <- payload:
		return nil
	case <-ps.ctx.Done():
		return ps.ctx.Err()
	}
}

func (ps *PubSub) Subscribe() error {
	if ps.handler == nil {
		return errors.Wrap(ErrInvalidPubSubConfig, "handler cannot be nil")
	}

	go func() {
		defer close(ps.ch)
		for {
			select {
			case msg := <-ps.ch:
				if err := ps.handler(msg); err != nil {
					ps.logger.Error("pubsub handler error", "topic", ps.topic, "error", err)
				}
			case <-ps.ctx.Done():
				return
			}
		}
	}()

	return nil
}

