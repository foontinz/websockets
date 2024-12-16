package sink

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisAddr = "localhost:6379"
const testChannel = "test"
const msgNumber = 10
const redisTestTimeout = time.Second * 5

func TestRedisClient(t *testing.T) {
	var (
		ctx, ctxCancel = context.WithTimeout(context.Background(), redisTestTimeout)
		receivedMsgs   = make(chan string, msgNumber)
		sentMsgs       = make(chan string, msgNumber)
	)
	defer func() {
		ctxCancel()
		close(receivedMsgs)
		close(sentMsgs)

	}()

	c := NewRedisClient(redis.NewClient(&redis.Options{
		Addr: redisAddr,
	}))
	defer c.Close(ctx)

	go func() {
		subChannel := c.Subscribe(ctx, testChannel)
		for msg := range subChannel {
			receivedMsgs <- msg
		}
	}()
	time.Sleep(time.Second)
	for i := 0; i < msgNumber; i++ {
		go func(i int) {
			msg := fmt.Sprintf("%d", i)
			if err := c.Write(ctx, testChannel, msg); err != nil {
				t.Errorf("Failed to publish %s, err: %v\n", msg, err)
			}
		}(i)
	}

	successfulCommunications := 0
	for successfulCommunications < msgNumber {
		select {
		case <-ctx.Done():
			t.FailNow()
		case <-receivedMsgs:
			successfulCommunications++
		}
	}
}
