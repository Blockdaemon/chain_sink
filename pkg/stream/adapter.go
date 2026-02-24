package stream

import "context"

type Adapter interface {
	HandleMessage(ctx context.Context, message []byte) error
}
