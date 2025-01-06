package events

type MessageEvent struct {
	Channel string `json:"channel"`
	Content []byte `json:"content"`
}
