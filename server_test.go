package client_test

import (
	"sync"
	"testing"
	"websocketReverseProxy/client"
	"websocketReverseProxy/server"
)

const clientAmount = 10
const msgAmount = 10
const serverAddr = ":8080"

func TestServer(t *testing.T) {
	go server.StartServer(serverAddr)

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
