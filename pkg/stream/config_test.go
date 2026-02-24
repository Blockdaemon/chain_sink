package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		cfg Config
		err error
	}{
		{cfg: Config{URL: "ws://localhost:8765/targets/ceac6435-b9af-441c-abf3-f78de9bfc32c/websocket", ApiKey: "123"}, err: nil},
		{cfg: Config{URL: "wss://localhost:8765/targets/ceac6435-b9af-441c-abf3-f78de9bfc32c/websocket", ApiKey: "123"}, err: nil},
		{cfg: Config{URL: "ws://localhost:8765/targets/ceac6435-b9af-441c-abf3-f78de9bfc32c/websockets", ApiKey: "123"}, err: ErrInvalidUrl},
		{cfg: Config{URL: "ws://localhost:8765/targets/invalid-id/websocket", ApiKey: "123"}, err: ErrInvalidTargetId},
		{cfg: Config{URL: "wss://svc.blockdaemon.com/targets/ceac6435-b9af-441c-abf3-f78de9bfc32c/websocket", ApiKey: "123"}, err: nil},
		{cfg: Config{URL: "wss://svc.blockdaemon.com/streaming/v2/targets/ceac6435-b9af-441c-abf3-f78de9bfc32c/websocket", ApiKey: ""}, err: ErrMissingApiKey},
	}

	for _, test := range tests {
		err := test.cfg.Validate()
		assert.ErrorIs(t, err, test.err)
	}
}
