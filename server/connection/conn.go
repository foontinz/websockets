package connection

import (
	"time"

	"github.com/gorilla/websocket"
)

func ConfigureConnection(conn *websocket.Conn) *websocket.Conn {
	// Set timeouts for read/write to avoid lingering connections (added this block)
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	return conn
}
