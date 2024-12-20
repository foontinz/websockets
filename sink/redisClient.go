package sink

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(client *redis.Client) *RedisClient {
	rc := &RedisClient{
		Client: client,
	}

	return rc
}

func (rc *RedisClient) Write(ctx context.Context, channel string, data interface{}) error {
	return rc.Client.Publish(ctx, channel, data).Err()
}

func (rc *RedisClient) Subscribe(ctx context.Context, channels ...string) (<-chan string, <-chan struct{}) {
	ch := make(chan string)
	ready := make(chan struct{})

	go func() {
		defer close(ch)
		pubsub := rc.Client.Subscribe(ctx, channels...)
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

func (rc *RedisClient) Close(ctx context.Context) error {
	return rc.Client.Close()
}
