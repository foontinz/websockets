package server

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	// Allow connections from any origin (not recommended for production)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func authenticateUpgrade(r *http.Request) bool {
	// left as placeholder for now
	return true
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Handler for each HTTP request to websocket endpoint
	// Authenticate the request and run goroutine to process it
	authenticateUpgrade(r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("SERVER: Error in upgrading:", err)
		return
	}
	defer conn.Close()
	for {
		messageType, message, err := conn.ReadMessage()
		if err == io.EOF {
			log.Println("SERVER: Got EOF ending")
			return
		}
		if err != nil {
			log.Println("SERVER: Error during reading message", err)
			continue
		}

		if messageType == websocket.BinaryMessage {
			log.Println("SERVER: Cannot process binary message")
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("SERVER: Error during echoing message:", err)
			continue
		}
		log.Printf("SERVER: Sent to client: %s\n", message)
	}
}

func StartServer(addr string) {
	http.HandleFunc("/ws", HandleConnections)
	log.Println("SERVER: WebSocket server starting on :8080")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("SERVER: Server failed: %v", err)
	}
}
