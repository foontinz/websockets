package serialization

import (
	"encoding/json"
	"websocketReverseProxy/pkg/events"
)

func DeserializeMessage(data []byte) (events.MessageEvent, error) {
	msgEvent := &events.MessageEvent{}
	if err := json.Unmarshal(data, msgEvent); err != nil {
		return *msgEvent, err
	}
	return *msgEvent, nil
}
