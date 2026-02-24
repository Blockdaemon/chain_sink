package stream

import (
	"regexp"

	"github.com/google/uuid"
)

type StreamMode string

const (
	StreamModeAck   StreamMode = "ack"
	StreamModeNoAck StreamMode = "noack"
)

type Config struct {
	Mode           StreamMode `mapstructure:"mode" default:"ack" validate:"oneof=ack noack"`
	URL            string     `mapstructure:"url" validate:"required"`
	Headers        []Header   `mapstructure:"headers"`
	WorkerPoolSize int        `mapstructure:"worker_pool_size" default:"1"`
	ApiKey         string     `mapstructure:"api_key"`
}

type Header struct {
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}

var urlRegex = regexp.MustCompile(`^(ws|wss):\/\/.*?\/targets\/(.*?)\/websocket$`)
var urlGatewayRegex = regexp.MustCompile(`^wss:\/\/(.+)?svc.blockdaemon.com\/streaming\/v2\/targets\/.*?\/websocket$`)

func (c *Config) Validate() error {
	matches := urlRegex.FindStringSubmatch(c.URL)
	if len(matches) != 3 {
		return ErrInvalidUrl
	}

	targetId := matches[2]
	_, err := uuid.Parse(targetId)
	if err != nil {
		return ErrInvalidTargetId
	}

	if urlGatewayRegex.MatchString(c.URL) && c.ApiKey == "" {
		return ErrMissingApiKey
	}

	return nil
}
