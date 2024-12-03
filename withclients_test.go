package client_test

import (
	"log"
	"sync"
	"testing"
	"websocketReverseProxy/client"
	"websocketReverseProxy/server"
)

const ClientAmount = 1
const serverAddr = ":8080"

func TestServer(t *testing.T) {
	go server.StartServer(serverAddr)

	wg := sync.WaitGroup{}
	var resChan = make(chan int, ClientAmount)
	for i := 0; i < ClientAmount; i++ {
		wg.Add(1)
		go client.RunClient(&wg, i, 1, resChan)
	}
	go func() {
		wg.Wait()      // Wait for all goroutines to finish
		close(resChan) // Close the channel
	}()

	// Read from the channel until it is closed
	for value := range resChan {
		log.Printf("TEST: Client successed %d\n", value)
		if value != 0 {
			t.Fail()
		}
	}

}
