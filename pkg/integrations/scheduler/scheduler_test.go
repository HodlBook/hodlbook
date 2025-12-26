package scheduler

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestScheduler_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var count atomic.Int32
	s, err := New(
		WithContext(ctx),
		WithLogger(discardLogger),
		WithInterval(50*time.Millisecond),
		WithHandler(func() error {
			count.Add(1)
			return nil
		}),
	)
	assert.NoError(t, err)

	err = s.Start()
	assert.NoError(t, err)
}

func TestScheduler_TicksMultipleTimes(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var count atomic.Int32
	s, err := New(
		WithContext(ctx),
		WithLogger(discardLogger),
		WithInterval(10*time.Millisecond),
		WithHandler(func() error {
			count.Add(1)
			return nil
		}),
	)
	assert.NoError(t, err)

	err = s.Start()
	assert.NoError(t, err)

	time.Sleep(55 * time.Millisecond)
	assert.GreaterOrEqual(t, count.Load(), int32(3))
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	var count atomic.Int32
	s, err := New(
		WithContext(ctx),
		WithLogger(discardLogger),
		WithInterval(10*time.Millisecond),
		WithHandler(func() error {
			count.Add(1)
			return nil
		}),
	)
	assert.NoError(t, err)

	err = s.Start()
	assert.NoError(t, err)

	time.Sleep(25 * time.Millisecond)
	cancel()
	countAtCancel := count.Load()

	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, countAtCancel, count.Load(), "should not tick after cancel")
}

func TestScheduler_InvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{"no context", []Option{WithLogger(discardLogger), WithInterval(time.Second), WithHandler(func() error { return nil })}},
		{"no interval", []Option{WithLogger(discardLogger), WithContext(context.Background()), WithHandler(func() error { return nil })}},
		{"no handler", []Option{WithLogger(discardLogger), WithContext(context.Background()), WithInterval(time.Second)}},
		{"no logger", []Option{WithContext(context.Background()), WithInterval(time.Second), WithHandler(func() error { return nil })}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts...)
			assert.Error(t, err)
		})
	}
}
