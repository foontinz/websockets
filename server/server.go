package server

import (
	"context"
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

type Clients map[uuid.UUID]*websocket.Conn

type ProxyServer struct {
	mu       sync.RWMutex
	upgrader websocket.Upgrader
	sink     sink.Sink
	clients  Clients
}

func NewProxyServer(sink sink.Sink) *ProxyServer {
	s := &ProxyServer{
		sink:    sink,
		clients: make(Clients),
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

func (ps *ProxyServer) getClient(uuid uuid.UUID) *websocket.Conn {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.clients[uuid]
}

func (ps *ProxyServer) addClient(connection *websocket.Conn) uuid.UUID {
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

func (ps *ProxyServer) RedirectWebsocketMessage(message events.MessageEvent) error {
	log.Printf("Redirected to dest message: %s\n", message)
	return nil
}

func (ps *ProxyServer) AcknowledgeClient(ctx context.Context, clientUUID uuid.UUID, message events.MessageEvent) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		conn := ps.getClient(clientUUID)
		if err := conn.WriteMessage(websocket.TextMessage, message.Content); err != nil {
			log.Println("SERVER: Error acknowledging client. Lost message:", err)
			return err
		}

		log.Printf("SERVER: Acknowledged client: %s\n", message.Content)
		return nil
	}
}

func (ps *ProxyServer) HandleWebsocketConnection(conn *websocket.Conn) {
	conn = connection.ConfigureConnection(conn)
	clientUUID := ps.addClient(conn)
	defer func() {
		log.Printf("Connection is closing %s, removing client from the server", clientUUID)
		ps.removeClient(clientUUID)
		conn.Close()
	}()
	for {
		ctx, ctxCancel := context.WithCancel(context.Background())
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("SERVER: Unexpected error:", err)
			}
			ctxCancel()
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
		ps.RedirectWebsocketMessage(msgEvent)
		ps.AcknowledgeClient(ctx, clientUUID, msgEvent)
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
