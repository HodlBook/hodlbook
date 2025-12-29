package wmPubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPubSub_PublishAndConsume(t *testing.T) {
	ch := make(chan []byte, 1)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	received := make(chan []byte, 1)
	sub := New(
		WithChannel(ch),
		WithContext(ctx),
		WithTopic("test-topic"),
		WithHandler(func(msg []byte) error {
			received <- msg
			return nil
		}),
	)
	err := sub.Subscribe()
	assert.NoError(t, err)

	pub := New(WithChannel(ch), WithContext(ctx), WithTopic("test-topic"))
	payload := []byte("hello world")
	err = pub.Publish(payload)
	assert.NoError(t, err)

	select {
	case msg := <-received:
		assert.Equal(t, payload, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive message in time")
	}
}

func TestPubSub_ContextCancellation(t *testing.T) {
	ch := make(chan []byte)
	ctx, cancel := context.WithCancel(t.Context())

	pub := New(WithChannel(ch), WithContext(ctx), WithTopic("test-topic"))

	cancel()

	err := pub.Publish([]byte("should fail"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPubSub_SubscribeWithoutHandler(t *testing.T) {
	ch := make(chan []byte, 1)

	sub := New(WithChannel(ch), WithContext(t.Context()), WithTopic("test-topic"))
	err := sub.Subscribe()
	assert.Error(t, err)
}
