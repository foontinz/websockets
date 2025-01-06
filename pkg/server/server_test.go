package server

import (
	"context"
	"sync"
	"testing"
	"websocketReverseProxy/pkg/client"
)

const clientAmount = 100
const msgAmount = 100
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

func TestServer(t *testing.T) {
	testSink := NewTestSink()
	go StartServer(serverAddr, testSink)

	wg := sync.WaitGroup{}
	var resChan = make(chan bool, clientAmount)
	for i := 0; i < clientAmount; i++ {
		wg.Add(1)
		go client.RunClient(&wg, i, msgAmount, resChan)
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
}
