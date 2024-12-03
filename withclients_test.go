package client_test

import (
	"log"
	"sync"
	"testing"
	"websocketReverseProxy/client"
)

const ClientAmount = 1

func TestServer(t *testing.T) {
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
		log.Printf("Client successed %d\n", value)
	}

}
