package events

import "websocketReverseProxy/pkg/sink"

type MessageEvent struct {
	Channel sink.Channel `json:"channel"`
	Content []byte       `json:"content"`
}
