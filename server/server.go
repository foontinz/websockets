package server

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
	"websocketReverseProxy/server/auth"
	"websocketReverseProxy/server/connection"
	"websocketReverseProxy/sink"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type InConnection struct {
	conn *websocket.Conn
}

type ProxyServer struct {
	mu       sync.RWMutex
	upgrader websocket.Upgrader
	sink     sink.Sink
	clients  map[uuid.UUID]InConnection
}

func NewProxyServer(sink sink.Sink) *ProxyServer {
	s := &ProxyServer{
		sink:    sink,
		clients: make(map[uuid.UUID]InConnection),
	}

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
	if !auth.AuthenticateUpgrade(r) {
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

func (ps *ProxyServer) addClient(connection InConnection) uuid.UUID {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	userUUID := uuid.New()
	ps.clients[userUUID] = connection
	return userUUID
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

	conn = connection.ConfigureConnection(conn)
	ps.addClient(InConnection{conn: conn})

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("SERVER: Unexpected error:", err)
			}
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

func StartServer(addr string, sink sink.Sink) {
	server := NewProxyServer(sink)
	http.HandleFunc("/ws", server.HandleConnections)
	log.Println("SERVER: WebSocket server starting on :8080")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("SERVER: Server failed: %v", err)
	}
}
