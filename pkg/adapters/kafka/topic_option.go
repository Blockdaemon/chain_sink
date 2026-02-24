package kafka

import (
	"fmt"
	"strconv"
	"time"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"go.uber.org/zap"
)

const (
	topicOptRetention   = "retention.ms"
	topicOptCompression = "compression.type"
)

type TopicCompressionType string

const (
	TopicCompressionTypeNone     TopicCompressionType = "uncompressed"
	TopicCompressionTypeProducer TopicCompressionType = "producer"
	TopicCompressionTypeGzip     TopicCompressionType = "gzip"
	TopicCompressionTypeSnappy   TopicCompressionType = "snappy"
	TopicCompressionTypeLz4      TopicCompressionType = "lz4"
	TopicCompressionTypeZstd     TopicCompressionType = "zstd"
)

// TopicOption is a function that alters a map of configuration. A few
// commonly used options are already provided. You can also write your own.
// See for reference the topic configuration reference:
// https://docs.confluent.io/platform/current/installation/configuration/topic-configs.html
type TopicOption func(map[string]string)

// Set topic retention period
func SetTopicRetention(timeFormat string) TopicOption {

	if timeFormat == "" {
		return func(configMap map[string]string) {
		}
	}

	if timeFormat == "-1" {
		return func(configMap map[string]string) {
			configMap[topicOptRetention] = "-1"
		}
	}

	duration, err := strconv.ParseInt(timeFormat, 10, 64)
	if err != nil {
		timeDuration, err := time.ParseDuration(timeFormat)
		if err != nil {
			logger.Log.Error("failed to parse topic retention", zap.Error(err), zap.String("time_format", timeFormat))
			return nil
		}
		duration = timeDuration.Milliseconds()
	}

	return func(configMap map[string]string) {
		configMap[topicOptRetention] = fmt.Sprintf("%d", duration)
	}
}

// Set topic compression type
func SetTopicCompressionType(compressionType TopicCompressionType) TopicOption {
	return func(configMap map[string]string) {
		if compressionType != "" {
			configMap[topicOptCompression] = string(compressionType)
		}
	}
}
