package sink

import (
	"context"
)

type Sink interface {
	Write(ctx context.Context, channels string, data interface{}) error
	Subscribe(ctx context.Context, channels ...string) <-chan string
	Close(ctx context.Context) error
}
