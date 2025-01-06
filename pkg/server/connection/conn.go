package connection

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

func ConfigureConnection(conn *websocket.Conn, ctxCancel context.CancelFunc) *websocket.Conn {
	// Set timeouts for read/write to avoid lingering connections (added this block)
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetCloseHandler(func(code int, text string) error {
		ctxCancel()
		return conn.CloseHandler()(code, text)
	})
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	return conn
}
