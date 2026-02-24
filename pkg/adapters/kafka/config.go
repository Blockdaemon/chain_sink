package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	clientOptBootstrapServers = "bootstrap.servers"

	clientOptSaslPassword     = "sasl.password"
	clientOptSaslUsername     = "sasl.username"
	clientOptSaslMechanism    = "sasl.mechanism"
	clientOptSecurityProtocol = "security.protocol"

	optSaslMechanismPlain = "PLAIN"

	optSecurityProtocolSSL     = "SSL"
	optSecurityProtocolSASLSSL = "SASL_SSL"

	clientOptBrokerAddressFamily         = "broker.address.family"
	clientOptPartitionAssignmentStrategy = "partition.assignment.strategy"
	clientOptGroupID                     = "group.id"
	clientOptAutoOffsetReset             = "auto.offset.reset"
	clientOptEnableAutoOffsetStore       = "enable.auto.offset.store"
	clientOptEnableAutoCommit            = "enable.auto.commit"
)

type AuthenticationType string

const (
	AuthenticationTypeNone AuthenticationType = "none"
	AuthenticationTypeSasl AuthenticationType = "sasl_ssl"
)

type Authentication struct {
	Type     AuthenticationType `mapstructure:"type" validate:"oneof='' none sasl_ssl" default:"none"`
	Username string             `mapstructure:"username"`
	Password string             `mapstructure:"password"`
}

func (a *Authentication) BuildOptions() ([]ClientOption, error) {
	switch a.Type {
	case AuthenticationTypeNone, "":
		return nil, nil
	case AuthenticationTypeSasl:
		return []ClientOption{UseSecurityProtocolSASLSSL(), UseSaslPlain(a.Username, a.Password)}, nil
	}
	return nil, fmt.Errorf("unsupported authentication type: %s", a.Type)
}

type ProducerConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type ClientOption func(kafka.ConfigMap) error

// Connect to Confluent Kafka using SSL
func UseSecurityProtocolSSL() ClientOption {
	return func(m kafka.ConfigMap) error {
		return m.SetKey(clientOptSecurityProtocol, optSecurityProtocolSSL)
	}
}

// Connect to Confluent Kafka using SASL SSL. Use this when
// SASL authentication is used.
func UseSecurityProtocolSASLSSL() ClientOption {
	return func(m kafka.ConfigMap) error {
		return m.SetKey(clientOptSecurityProtocol, optSecurityProtocolSASLSSL)
	}
}

// Connection to Confluent Kafka using SASL PLAIN authentication
func UseSaslPlain(username, password string) ClientOption {
	return func(m kafka.ConfigMap) error {
		if err := m.SetKey(clientOptSaslMechanism, optSaslMechanismPlain); err != nil {
			return err
		}
		if err := m.SetKey(clientOptSaslUsername, username); err != nil {
			return err
		}
		if err := m.SetKey(clientOptSaslPassword, password); err != nil {
			return err
		}
		return nil
	}
}

func WithAutoOffsetCommit() ClientOption {
	return func(m kafka.ConfigMap) error {
		return m.SetKey(clientOptEnableAutoCommit, "true")
	}
}
