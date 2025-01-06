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

func TestRedisSink(t *testing.T) {
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

	c := NewRedisSink(redis.NewClient(&redis.Options{
		Addr: redisAddr,
	}))
	defer c.Close(ctx)

	t.Log("Starting subscription...")
	subChannel, ready := c.Subscribe(ctx, testChannel)
	select {
	case <-ctx.Done():
		t.Fatal("Context canceled while waiting for subscription")
	case <-ready:
		go func() {
			for msg := range subChannel {
				receivedMsgs <- msg
				t.Logf("Received message: %s", msg)
			}
		}()
	}

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
			t.Fatalf("Timeout waiting for messages. Received only %d/%d",
				successfulCommunications, msgNumber)
		case <-receivedMsgs:
			successfulCommunications++
		}
	}
}
