package server

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
	"websocketReverseProxy/events"
	"websocketReverseProxy/serialization"
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

func (ps *ProxyServer) removeClient(uuid uuid.UUID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.clients, uuid)
}

func (ps *ProxyServer) HandleWebsocketMessage(conn *websocket.Conn, message events.MessageEvent) error {

	if err := conn.WriteMessage(websocket.TextMessage, message.Content); err != nil {
		log.Println("SERVER: Error during writing to client message:", err)
		return err
	}

	log.Printf("SERVER: Sent to client: %s\n", message.Content)
	return nil
}

func (ps *ProxyServer) HandleWebsocketConnection(conn *websocket.Conn) {
	defer conn.Close()
	conn = connection.ConfigureConnection(conn)
	connId := ps.addClient(InConnection{conn: conn})
	defer func() {
		log.Printf("Connection is closing %s, removing client from the server", connId)
		ps.removeClient(connId)
	}()

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

		msgEvent, err := serialization.DeserializeMessage(message)
		if err != nil {
			log.Println("SERVER: Cannot process message, not deserializable, msg: ", message)
			continue
		}
		ps.HandleWebsocketMessage(conn, msgEvent)
	}
}

func (ps *ProxyServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := ps.tryUpgradeConn(w, r)
	if err != nil {
		log.Println("SERVER: Failed to upgrade connection to websockets.")
		return
	}
	go ps.HandleWebsocketConnection(conn)
}

func StartServer(addr string, sink sink.Sink) {
	server := NewProxyServer(sink)
	http.HandleFunc("/ws", server.HandleConnections)
	log.Println("SERVER: WebSocket server starting on :8080")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("SERVER: Server failed: %v", err)
	}
}
