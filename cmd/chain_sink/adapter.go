package main

import (
	"context"
	"fmt"

	"github.com/blockdaemon/chain_sink/pkg/adapters/kafka"
	"github.com/blockdaemon/chain_sink/pkg/adapters/stdout"
	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/stream"
	"go.uber.org/zap"
)

func buildAdapter(ctx context.Context, cfg AdapterConfig) (stream.Adapter, error) {
	switch cfg.Type {
	case AdapterTypeStdout:
		return new(stdout.StdoutAdapter), nil
	case AdapterTypeKafka:
		if cfg.Kafka == nil {
			return nil, fmt.Errorf("kafka config is required")
		}

		opts, err := cfg.Kafka.Authentication.BuildOptions()
		if err != nil {
			return nil, fmt.Errorf("error building authentication options: %w", err)
		}

		if cfg.Kafka.CreateTopic {
			adminClient, err := kafka.NewAdminClient(cfg.Kafka.AdminHost, opts...)
			if err != nil {
				return nil, fmt.Errorf("error creating admin client: %w", err)
			}

			topicOptions := []kafka.TopicOption{
				kafka.SetTopicCompressionType(cfg.Kafka.CompressionType),
				kafka.SetTopicRetention(cfg.Kafka.RetentionTime),
			}

			if _, err := adminClient.CreateTopicIfNotExists(ctx, cfg.Kafka.TopicName, cfg.Kafka.NumPartitions, cfg.Kafka.ReplicationFactor, topicOptions...); err != nil {
				return nil, fmt.Errorf("error creating topic: %w", err)
			}
		}

		producer, err := kafka.NewKafkaAdapter(cfg.Kafka.Producer, opts...)
		if err != nil {
			return nil, fmt.Errorf("error creating producer: %w", err)
		}

		go func() {
			if err := producer.Run(ctx); err != nil {
				logger.Log.Error("error running producer", zap.Error(err))
			}
		}()

		return producer, nil
	}
	return nil, fmt.Errorf("unsupported adapter type: %s", cfg.Type)
}
