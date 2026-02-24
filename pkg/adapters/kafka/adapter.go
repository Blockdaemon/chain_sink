package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/stream"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var _ stream.Adapter = (*KafkaAdapter)(nil)

type KafkaAdapter struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaAdapter(cfg ProducerConfig, opts ...ClientOption) (*KafkaAdapter, error) {
	clientConfig := kafka.ConfigMap{
		clientOptBootstrapServers: strings.Join(cfg.Brokers, ","),
	}

	for _, opt := range opts {
		if err := opt(clientConfig); err != nil {
			return nil, fmt.Errorf("error applying client option: %w", err)
		}
	}

	producer, err := kafka.NewProducer(&clientConfig)
	if err != nil {
		return nil, err
	}

	return &KafkaAdapter{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

func (p *KafkaAdapter) Run(ctx context.Context) error {
	events := p.producer.Events()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-events:
			switch e := event.(type) {
			case *kafka.Message:
				if e.TopicPartition.Error != nil {
					return e.TopicPartition.Error
				}
				logger.Log.Debug("Message sent", zap.String("topic", *e.TopicPartition.Topic), zap.Int32("partition", e.TopicPartition.Partition), zap.Int64("offset", int64(e.TopicPartition.Offset)))

			case kafka.Error:
				// These are considered non-fatal errors which are retried automatically. Might drop to debug level later.
				logger.Log.Error("Kafka error", zap.Error(e))
			}
		}
	}
}

func (p *KafkaAdapter) HandleMessage(ctx context.Context, message []byte) error {
	deliveries := make(chan kafka.Event)

	// random uuid as key
	// TODO: reassess this strategy later
	key := uuid.New().String()

	if err := p.producer.Produce(&kafka.Message{
		Value:          message,
		Key:            []byte(key),
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
	}, deliveries); err != nil {
		return err
	}

	// WHAT IS GOING ON HERE?
	// since we are receiving from a request that may be cancelled via context we would like to return if the context is cancelled.
	// note that the kafka produce is async so produced messages may not be sent to the broker yet in the event of the app crashing.
	// so we need to wait for the message delivery message to be sure its been sent to the broker, however bad practice
	// but the kafka lib requires the reader to close the delivery channel not the writer.
	// so this channel needs to be kept open until a message is received. this routine is used to decouple the call from the context.
	delivered := make(chan error, 1)
	go func(ctx context.Context) {
		defer close(deliveries)
		defer close(delivered)

		select {
		case event := <-deliveries:
			switch e := event.(type) {
			case *kafka.Message:
				if e.TopicPartition.Error != nil {
					delivered <- e.TopicPartition.Error
				} else {
					return
				}
			case kafka.Error:
				delivered <- e
			default:
				delivered <- fmt.Errorf("unexpected event type %T", event)
			}
		case <-ctx.Done():
			select {
			case delivered <- ctx.Err():
			default:
				return
			}
		}
	}(ctx)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err2, ok := <-delivered:
		if !ok {
			return nil
		}
		if err2 == nil {
			logger.Log.Debug("Message delivered to kafka", zap.String("topic", p.topic), zap.String("key", key))
		}
		return err2
	}
}

func (p *KafkaAdapter) Close() error {
	if p.producer.IsClosed() {
		return nil
	}

	p.producer.Close()
	outstandingMessages := p.producer.Flush(30 * 1000)
	if outstandingMessages > 0 {
		logger.Log.Warn("Failed to send all messages", zap.Int("outstanding_messages", outstandingMessages))
	}

	return nil
}
