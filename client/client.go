package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	conn                *websocket.Conn
	NonSuccessfulEchoes uint8 // amount of successful echoes
}

func newClient(serverURL string) (*Client, error) {
	// Establish WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, http.Header{})
	if err != nil {
		return nil, fmt.Errorf("CLIENT: failed to connect: %v", err)
	}

	return &Client{
		conn:                conn,
		NonSuccessfulEchoes: 0,
	}, nil
}

func (c *Client) sendMessage(message string) {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("CLIENT: Send error:", err)
		return
	}
	c.NonSuccessfulEchoes++
}

func (c *Client) receiveMessage(expMessage string) {
	_, actMessage, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("CLIENT: Read error:", err)
		return
	}
	if expMessage != string(actMessage) {
		c.NonSuccessfulEchoes--
	}
}

func (c *Client) close() {
	c.conn.Close()
}

func RunClient(wg *sync.WaitGroup, clientNum int, msgNum int, result chan<- int) {
	// Goroutine to run client
	defer wg.Done()

	client, err := newClient("ws://localhost:8080/ws")
	if err != nil {
		log.Printf("Client creation failed: %v\n", err)
		return
	}
	defer client.close()

	for i := 0; i < msgNum; i++ {
		msg := fmt.Sprintf("MESSAGE: Client: #%d, my %d message", clientNum, i)
		client.sendMessage(msg)
		client.receiveMessage(msg)
	}

	log.Printf("CLIENT: Client: #%d, ive sent all msgs\n", clientNum)

	if client.NonSuccessfulEchoes == 0 {
		result <- 1
	} else {
		result <- 0
	}
}
