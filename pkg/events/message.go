package events

import "websocketReverseProxy/pkg/sink"

type MessageEvent struct {
	Channel sink.Channel `json:"channel"`
	Content []byte       `json:"content"`
}

func NewMessageEvent(channel sink.Channel, content []byte) *MessageEvent {
	return &MessageEvent{
		Channel: channel,
		Content: content,
	}
}
