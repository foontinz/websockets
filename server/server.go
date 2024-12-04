package server

import (
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time" // Added for timeout handling
)

func authenticateUpgrade(r *http.Request) bool {
	// left as placeholder for now
	return true
}

type ProxyServer struct {
	upgrader websocket.Upgrader
}

func NewProxyServer() *ProxyServer {
	s := &ProxyServer{}
	s.upgrader = websocket.Upgrader{
		HandshakeTimeout: time.Second * 30,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return s
}

func (ps *ProxyServer) tryUpgradeConn(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	if !authenticateUpgrade(r) {
		log.Println("SERVER: Failed to authenticate upgrade")
		return nil, errors.New("authentication error")
	}

	conn, err := ps.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("SERVER: Errored in upgrading:", err)
		return nil, err
	}
	return conn, nil
}
func (ps *ProxyServer) configureConnection(conn *websocket.Conn) *websocket.Conn {
	// Set timeouts for read/write to avoid lingering connections (added this block)
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	return conn
}

func (ps *ProxyServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := ps.tryUpgradeConn(w, r)
	if err != nil {
		log.Println("SERVER: Failed to upgrade connection to websockets.")
		return
	}
	defer func() {
		log.Println("SERVER: Closing connection")
		conn.Close()
	}()
	conn = ps.configureConnection(conn)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("SERVER: Unexpected close error:", err)
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				log.Println("SERVER: Unexpected eof error:", err)
				break
			}
			log.Println("SERVER: Error during reading message:", err)
			break
		}

		if messageType == websocket.BinaryMessage {
			log.Println("SERVER: Cannot process binary message")
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("SERVER: Error during echoing message:", err)
			return
		}
		log.Printf("SERVER: Sent to client: %s\n", message)
	}
}

func StartServer(addr string) {
	server := NewProxyServer()
	http.HandleFunc("/ws", server.HandleConnections)
	log.Println("SERVER: WebSocket server starting on :8080")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("SERVER: Server failed: %v", err)
	}
}
