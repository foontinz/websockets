package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"reflect"
	"sync"
)

type Client struct {
	conn             *websocket.Conn
	sentMessages     []string
	receivedMessages []string
}

func newClient(serverURL string) (*Client, error) {
	// Establish WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, http.Header{})
	if err != nil {
		return nil, fmt.Errorf("CLIENT: failed to connect: %v", err)
	}

	return &Client{
		conn:             conn,
		sentMessages:     make([]string, 0, 32),
		receivedMessages: make([]string, 0, 32),
	}, nil
}

func (c *Client) sendMessage(message string) {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("CLIENT: Send error:", err)
		return
	}
	c.sentMessages = append(c.sentMessages, message)
}

func (c *Client) receiveMessage() {
	_, actMessage, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("CLIENT: Read error:", err)
		return
	}
	c.receivedMessages = append(c.receivedMessages, string(actMessage))
}

func (c *Client) close() {
	c.conn.Close()
}

func RunClient(wg *sync.WaitGroup, clientNum int, msgNum int, result chan<- bool) {
	// Goroutine to run client
	defer wg.Done()

	client, err := newClient("ws://localhost:8080/ws")
	if err != nil {
		log.Printf("Client creation failed: %v\n", err)
		return
	}
	defer client.close()

	for i := 0; i < msgNum; i++ {
		msg := fmt.Sprintf("i am client: #%d, my %d message", clientNum, i)
		client.sendMessage(msg)
		client.receiveMessage()
	}
	result <- reflect.DeepEqual(client.receivedMessages, client.sentMessages)
}
