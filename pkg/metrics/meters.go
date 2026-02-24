package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

type Meters interface {
	RecordMessagesReceived(ctx context.Context)
	RecordMessagesAcked(ctx context.Context)
	RecordMessagesForwardedToAdapter(ctx context.Context)
}

type OtelMeters struct {
	messagesReceived           metric.Int64Counter
	messagesAcked              metric.Int64Counter
	messagesForwardedToAdapter metric.Int64Counter
}

func New(provider metric.MeterProvider) (*OtelMeters, error) {
	meter := provider.Meter("com.blockdaemon.chain_sink")

	messagesReceived, err := meter.Int64Counter("messages_received")
	if err != nil {
		return nil, err
	}

	messagesAcked, err := meter.Int64Counter("messages_acked")
	if err != nil {
		return nil, err
	}

	messagesForwardedToAdapter, err := meter.Int64Counter("messages_forwarded_to_adapter")
	if err != nil {
		return nil, err
	}

	return &OtelMeters{
		messagesReceived:           messagesReceived,
		messagesAcked:              messagesAcked,
		messagesForwardedToAdapter: messagesForwardedToAdapter,
	}, nil
}

var _ Meters = (*OtelMeters)(nil)

func (m *OtelMeters) RecordMessagesReceived(ctx context.Context) {
	m.messagesReceived.Add(ctx, 1)
}

func (m *OtelMeters) RecordMessagesAcked(ctx context.Context) {
	m.messagesAcked.Add(ctx, 1)
}

func (m *OtelMeters) RecordMessagesForwardedToAdapter(ctx context.Context) {
	m.messagesForwardedToAdapter.Add(ctx, 1)
}
