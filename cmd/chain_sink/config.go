package main

import (
	"github.com/blockdaemon/chain_sink/pkg/adapters/kafka"
	"github.com/blockdaemon/chain_sink/pkg/config"
	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/stream"
)

type Config struct {
	Logger      logger.Config `mapstructure:"logger"`
	Stream      stream.Config `mapstructure:"stream"`
	StreamCount int           `mapstructure:"stream_count" default:"1"`
	Adapter     AdapterConfig `mapstructure:"adapter"`
	Metrics     MetricsConfig `mapstructure:"metrics"`
}

type AdapterType string

const (
	AdapterTypeStdout AdapterType = "stdout"
	AdapterTypeKafka  AdapterType = "kafka"
)

type AdapterConfig struct {
	Type  AdapterType  `mapstructure:"type" validate:"oneof=stdout kafka" default:"stdout"`
	Kafka *KafkaConfig `mapstructure:"kafka"`
}

type KafkaConfig struct {
	Authentication    kafka.Authentication       `mapstructure:"authentication"`
	Producer          kafka.ProducerConfig       `mapstructure:"producer" validate:"required"`
	CreateTopic       bool                       `mapstructure:"create_topic"`
	TopicName         string                     `mapstructure:"topic_name" validate:"required"`
	NumPartitions     int                        `mapstructure:"num_partitions" default:"1"`
	ReplicationFactor int                        `mapstructure:"replication_factor" default:"1"`
	CompressionType   kafka.TopicCompressionType `mapstructure:"compression_type" default:"uncompressed" validate:"oneof='' uncompressed producer gzip snappy lz4 zstd"`
	RetentionTime     string                     `mapstructure:"retention_time"`
	AdminHost         string                     `mapstructure:"admin_host"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port" default:"8421"`
}

func LoadConfig() (Config, error) {
	return config.LoadConfig[Config]()
}
