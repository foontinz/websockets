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

func (rc *RedisClient) Subscribe(ctx context.Context, channels ...string) <-chan string {
	ch := make(chan string)
	go func() {
		for msg := range rc.Client.Subscribe(ctx, channels...).Channel() {
			ch <- msg.Payload
		}
	}()
	return ch
}

func (rc *RedisClient) Close(ctx context.Context) error {
	return rc.Client.Close()
}
