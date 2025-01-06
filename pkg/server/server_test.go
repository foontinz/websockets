package server

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"websocketReverseProxy/pkg/dummy_client"
	"websocketReverseProxy/pkg/events"
	"websocketReverseProxy/pkg/sink"
)

const serverAddr = ":8080"

type TestSink struct {
	mu         sync.Mutex
	redirected map[string][]interface{}
}

func NewTestSink() *TestSink {
	return &TestSink{
		redirected: make(map[string][]interface{}),
		mu:         sync.Mutex{},
	}
}

func (ts *TestSink) Write(ctx context.Context, channel string, data interface{}) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.redirected[channel] = append(ts.redirected[channel], data)
	return nil
}

func (ts *TestSink) Close(ctx context.Context) error {
	return nil
}

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomMessages(maxMsgSize int, msgAmount int) []events.MessageEvent {
	messages := make([]events.MessageEvent, msgAmount, msgAmount)
	for i := 0; i < msgAmount; i++ {
		msgSize := rand.Intn(maxMsgSize)
		messages[i] = *events.NewMessageEvent(sink.Channel(RandomString(msgSize)), []byte(RandomString(msgSize)))
	}
	return messages
}

func TestServer(t *testing.T) {
	tests := []struct {
		name          string
		clientsAmount int
		input         []events.MessageEvent
	}{
		{"Positive 1 client, 1 message with up to 10 chars of content", 1, RandomMessages(10, 1)},
		{"Positive 10 client, 10 message with 10 chars of content", 10, RandomMessages(10, 10)},
		{"Positive 100 client, 100 message with 100 chars of content", 100, RandomMessages(100, 100)},
	}
	testSink := NewTestSink()
	go StartServer(serverAddr, testSink)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := sync.WaitGroup{}
			var resChan = make(chan bool, tt.clientsAmount)
			for i := 0; i < tt.clientsAmount; i++ {
				wg.Add(1)
				go dummy_client.RunClient(&wg, tt.input, resChan)
			}

			go func() {
				wg.Wait()
				close(resChan)
			}()

			for value := range resChan {
				if !value {
					t.Fail()
				}
			}
		})
	}
}
