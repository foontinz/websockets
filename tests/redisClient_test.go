package test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisAddr = "localhost:6379"
const testChannel = "test"
const msgNumber = 100

func subscribeOnChannel(ctx context.Context, c *redis.Client, channel string) {
	pubsub := c.Subscribe(ctx, testChannel)
	ch := pubsub.Channel()
	for msg := range ch {
		log.Printf("From ch: %s, read msg: %s\n", channel, msg.String())
	}
}

const RedisTimeout = time.Second * 15

func TestRedisClient(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer ctxCancel()
	c := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	go subscribeOnChannel(ctx, c, testChannel)
	for i := 0; i < msgNumber; i++ {
		c.Publish(ctx, testChannel, fmt.Sprintf("%d", i))
	}
}
