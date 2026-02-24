package kafka

import (
	"context"
	"fmt"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Administrator interface {
	CreateTopicIfNotExists(ctx context.Context, topicName string, numPartitions, replicationFactor int, opts ...TopicOption) (bool, error)
}

var _ Administrator = (*AdminClient)(nil)

// ConfluentAdmin implements Administrator interface and allows to create topics by code
// NOTE: In order to create a topic with this function you need both `DESCRIBE` and `CREATE` ACL privileges on either the cluster level, or on the topic level.
type AdminClient struct {
	client *kafka.AdminClient
}

// NewAdminClient creates a new Confluent admin client. To use this with confluent cloud use:
//
//	NewAdminClient("https://host.to.confluent:1234", UseSaslPlain("KEY_ID", "KEY_SECRET"), UseSecurityProtocolSASLSSL())
func NewAdminClient(host string, opts ...ClientOption) (*AdminClient, error) {
	confluentOpts := kafka.ConfigMap{
		clientOptBootstrapServers: host,
	}

	for _, opt := range opts {
		if err := opt(confluentOpts); err != nil {
			return nil, err
		}
	}

	client, err := kafka.NewAdminClient(&confluentOpts)
	if err != nil {
		return nil, err
	}

	return &AdminClient{
		client: client,
	}, nil
}

// CreateTopicIfNotExists creates a topic if it does not exist yet. This will return true if the topic is created. If it returns false and no error the topic
// already existed and no operation was performed.
// NOTE: In order to create a topic with this function you need both `DESCRIBE` and `CREATE` ACL privileges on either the cluster level, or on the topic level.
func (k *AdminClient) CreateTopicIfNotExists(ctx context.Context, topicName string, numPartitions, replicationFactor int, opts ...TopicOption) (bool, error) {
	meta, err := k.client.GetMetadata(&topicName, false, 60000)
	if err != nil {
		return false, fmt.Errorf("error retrieving topic metadata from confluent: %w", err)
	}

	if result, ok := meta.Topics[topicName]; ok {
		if result.Error.Code() == kafka.ErrNoError {
			return false, nil
		}

		if result.Error.Code() != kafka.ErrUnknownTopic && result.Error.Code() != kafka.ErrUnknownTopicOrPart {
			return false, fmt.Errorf("failed to query topic. Returned code %d: %w", result.Error.Code(), result.Error)
		}
	}

	topicConfig := make(map[string]string)

	for _, opt := range opts {
		opt(topicConfig)
	}

	logger.Log.Info("creating kafka topic", zap.String("topic", topicName), zap.Int("num_partitions", numPartitions), zap.Int("replication_factor", replicationFactor), zap.Any("config", topicConfig))

	result, err := k.client.CreateTopics(ctx, []kafka.TopicSpecification{{
		Topic:             topicName,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
		Config:            topicConfig,
	}})
	if err != nil {
		return false, fmt.Errorf("error creating kafka topic %s: %w", topicName, err)
	}

	for _, topicResult := range result {
		if topicResult.Error.Code() != kafka.ErrNoError && topicResult.Error.Code() != kafka.ErrTopicAlreadyExists {
			return false, fmt.Errorf("result: error creating kafka topic %s: %w", topicResult.Topic, topicResult.Error)
		}
	}

	return true, nil
}
