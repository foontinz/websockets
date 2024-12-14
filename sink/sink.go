package sink

import "context"

type Sink interface {
	Write(ctx context.Context, data []byte) error
	Read(ctx context.Context) ([]byte, error)
}
