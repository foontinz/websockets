package test

import (
	"context"
	"fmt"
	"log"
	"testing"

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

func TestRedisClient(t *testing.T) {
	c := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	go subscribeOnChannel(context.TODO(), c, testChannel)
	for i := 0; i < msgNumber; i++ {
		c.Publish(context.TODO(), testChannel, fmt.Sprintf("%d", i))
	}
}
