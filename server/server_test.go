package server

import (
	"sync"
	"testing"
	"websocketReverseProxy/client"
	"websocketReverseProxy/sink"
)

const clientAmount = 100
const msgAmount = 100
const serverAddr = ":8080"

func TestServer(t *testing.T) {
	go StartServer(serverAddr, sink.Sink(nil))

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
