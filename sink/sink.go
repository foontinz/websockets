package sink

import (
	"context"
)

type Sink interface {
	Write(ctx context.Context, channels string, data interface{}) error
	// Subscribe(ctx context.Context, channels ...string) (<-chan string, <-chan struct{}) TODO: might be extended in the future
	Close(ctx context.Context) error
}
