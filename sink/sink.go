package sink

import (
	"context"
)

type Sink interface {
	Subscribe(ctx context.Context, channels ...string) <-chan string
	Read(ctx context.Context, channel string) ([]byte, error)
	Close(ctx context.Context) error
}
