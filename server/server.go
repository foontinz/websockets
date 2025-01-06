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
		log.Println("[SERVER]: Failed to authenticate upgrade")
		return nil, errors.New("[SERVER]: authentication error")
	}

	conn, err := ps.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[SERVER]: Errored in upgrading:", err)
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

func (ps *ProxyServer) RedirectWebsocketMessage(clientUUID uuid.UUID, message events.MessageEvent) error {
	if err := ps.sink.Write(context.TODO(), clientUUID.String(), message.Content); err != nil {
		return err
	}
	log.Printf("[SERVER]: Redirected to dest message: %s\n", message)
	return nil
}

func (ps *ProxyServer) AcknowledgeClient(ctx context.Context, clientUUID uuid.UUID, message events.MessageEvent) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		conn := ps.getClient(clientUUID)
		if err := conn.WriteMessage(websocket.TextMessage, message.Content); err != nil {
			return err
		}
		log.Printf("[SERVER]: Acknowledged client: %s\n", message.Content)
		return nil
	}
}

func (ps *ProxyServer) HandleWebsocketConnection(conn *websocket.Conn) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	conn = connection.ConfigureConnection(conn, ctxCancel)
	clientUUID := ps.addClient(conn)

	defer func() {
		log.Printf("[SERVER]: Connection is closing %s, removing client from the server", clientUUID)
		ctxCancel()
		ps.removeClient(clientUUID)
		conn.Close()
	}()

Loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("[SERVER]: Connection is closed or cancelled")
			return
		default:
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("[SERVER]: Unexpected error:", err)
				}
				break Loop
			}
			if messageType == websocket.BinaryMessage {
				log.Println("[SERVER]: Cannot process binary message")
				continue Loop
			}

			msgEvent, err := serialization.DeserializeMessage(message)
			if err != nil {
				log.Println("[SERVER]: Cannot process message, not deserializable, msg: ", message)
				continue Loop
			}

			if err := ps.RedirectWebsocketMessage(clientUUID, msgEvent); err != nil {
				log.Printf("[SERVER]: Error redirecting websocket message: %s\n", message)
				continue Loop
			}
			if err := ps.AcknowledgeClient(ctx, clientUUID, msgEvent); err != nil {
				log.Println("[SERVER]: Error acknowledging client. Lost message:", err)
			}
		}
	}
}

func (ps *ProxyServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := ps.tryUpgradeConn(w, r)
	if err != nil {
		log.Println("[SERVER]: Failed to upgrade connection to websockets.")
		return
	}
	go ps.HandleWebsocketConnection(conn)
}

func StartServer(addr string, sink sink.Sink) {
	server := NewProxyServer(sink)
	http.HandleFunc("/ws", server.HandleConnections)
	log.Println("[SERVER]: WebSocket server starting on :8080")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("[SERVER]: Server failed: %v", err)
	}
}
