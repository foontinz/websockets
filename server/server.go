package server

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time" // Added for timeout handling
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
	authenticateUpgrade(r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("SERVER: Error in upgrading:", err)
		return
	}

	defer func() {
		log.Println("SERVER: Closing connection")
		conn.Close()
	}()

	// Set timeouts for read/write to avoid lingering connections (added this block)
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, message, err := conn.ReadMessage()
		if err == io.EOF {
			log.Println("SERVER: Got EOF ending")
			return
		}
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("SERVER: Unexpected close error:", err)
			return
		}
		if err != nil {
			log.Println("SERVER: Error during reading message:", err)
			return // Modified to close the loop on error
		}

		if messageType == websocket.BinaryMessage {
			log.Println("SERVER: Cannot process binary message")
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("SERVER: Error during echoing message:", err)
			return // Modified to close the loop on write error
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
