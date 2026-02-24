package stdout

import (
	"context"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/stream"
	"go.uber.org/zap"
)

var _ stream.Adapter = (*StdoutAdapter)(nil)

type StdoutAdapter struct{}

func (StdoutAdapter) HandleMessage(_ context.Context, message []byte) error {
	logger.Log.Info("received message", zap.String("message", string(message)))
	return nil
}
