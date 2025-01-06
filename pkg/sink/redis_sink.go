package sink

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisSink struct {
	client *redis.Client
}

func NewRedisSink(client *redis.Client) *RedisSink {
	rc := &RedisSink{
		client: client,
	}

	return rc
}

func (rc *RedisSink) Write(ctx context.Context, channel string, data interface{}) error {
	return rc.client.Publish(ctx, channel, data).Err()
}

func (rc *RedisSink) Subscribe(ctx context.Context, channels ...string) (<-chan string, <-chan struct{}) {
	ch := make(chan string)
	ready := make(chan struct{})

	go func() {
		defer close(ch)
		pubsub := rc.client.Subscribe(ctx, channels...)
		defer pubsub.Close()
		ready <- struct{}{}
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pubsub.Channel():
				ch <- msg.Payload
			}
		}
	}()
	return ch, ready
}

func (rc *RedisSink) Close(ctx context.Context) error {
	return rc.client.Close()
}
