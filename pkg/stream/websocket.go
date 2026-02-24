package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/metrics"
	"github.com/coder/websocket"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidUrl      = errors.New("invalid url: make sure the url is in the format wss://<host>:<port>/targets/<target_id>/websocket")
	ErrInvalidTargetId = errors.New("invalid target id: the url is valid, but the target id is not a valid uuid")
	ErrMissingApiKey   = errors.New("API key is required: please specify the API key to authenticate with the Chain Watch API")
)

type ChainWatchStream struct {
	sync.RWMutex
	cfg Config

	conn     *websocket.Conn
	workChan chan []byte
}

func NewChainWatchStream(ctx context.Context, cfg Config) (*ChainWatchStream, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	stream := &ChainWatchStream{
		cfg:      cfg,
		workChan: make(chan []byte, cfg.WorkerPoolSize),
	}

	if err := stream.establishConnection(ctx); err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *ChainWatchStream) establishConnection(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()

	if s.conn != nil {
		_ = s.conn.Close(websocket.StatusNormalClosure, "reestablishing connection")
	}

	url := s.cfg.URL
	headers := make(http.Header)

	if s.cfg.ApiKey != "" {
		headers.Add("x-api-key", s.cfg.ApiKey)
	}

	for _, header := range s.cfg.Headers {
		headers.Add(header.Key, header.Value)
	}

	conn, resp, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		var body []byte
		if resp != nil && resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
			defer resp.Body.Close()
		}
		logger.Log.Error("error dialing websocket", zap.Error(err), zap.String("url", url), zap.Any("headers", headers), zap.String("response", string(body)), zap.Int("status", resp.StatusCode))
		return err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	s.conn = conn
	return nil
}

// reestablishes the connection if the websocket error is a closed error. If err is nil, it will reestablish the connection as well.
func (s *ChainWatchStream) reestablishConnection(ctx context.Context, err error) error {
	shouldReestablish := false
	if err == nil {
		shouldReestablish = true
	}

	var closeError *websocket.CloseError
	if errors.As(err, &closeError) {
		shouldReestablish = true
	}

	if errors.Is(err, io.EOF) {
		shouldReestablish = true
	}

	if !shouldReestablish {
		return err
	}

	logger.Log.Warn("reestablishing connection after websocket close", zap.Error(err))
	return s.establishConnection(ctx)
}

func (s *ChainWatchStream) ForwardMessagesToAdapter(ctx context.Context, adapter Adapter) error {
	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return s.readFromWebsocket(gCtx)
	})

	logger.Log.Debug("starting worker pool", zap.Int("worker_pool_size", s.cfg.WorkerPoolSize))
	for range s.cfg.WorkerPoolSize {
		group.Go(func() error {
			return s.runWorker(gCtx, adapter)
		})
	}

	return group.Wait()
}

func (s *ChainWatchStream) readFromWebsocket(ctx context.Context) error {
	for {
		// Message type is always text or binary, so we don't need to check it.
		// pings, pongs and closures are handled by the websocket library itself.
		_, message, err := s.conn.Read(ctx)
		if err != nil {
			logger.Log.Error("error reading from websocket", zap.Error(err))
			if err := s.reestablishConnection(ctx, err); err != nil {
				return err
			}
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case s.workChan <- message:
			metrics.G.RecordMessagesReceived(ctx)
		}
	}
}

func (s *ChainWatchStream) runWorker(ctx context.Context, adapter Adapter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message := <-s.workChan:
			if err := s.handleMessage(ctx, message, adapter); err != nil {
				return err
			}
			metrics.G.RecordMessagesForwardedToAdapter(ctx)
		}
	}
}

var parserPool = fastjson.ParserPool{}
var arenaPool = fastjson.ArenaPool{}

var bytesPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

func (s *ChainWatchStream) handleMessage(ctx context.Context, message []byte, adapter Adapter) error {

	// NoAck mode does not require any acknowledgement, so we can just forward the message to the adapter.
	// this is much faster and simpler than ack mode, but less resilient. If the adapter fails the message will be lost.
	if s.cfg.Mode == StreamModeNoAck {
		return adapter.HandleMessage(ctx, message)
	}

	// Ack mode requires an acknowledgement, so we parse the message ID and only send it back as acknowledgement
	// when the message is successfully handled by the adapter.
	parser := parserPool.Get()
	defer parserPool.Put(parser)

	parsed, err := parser.ParseBytes(message)
	if err != nil {
		return err
	}

	messageId := parsed.Get("id")
	if messageId == nil {
		return fmt.Errorf("message id is required")
	}

	if err := adapter.HandleMessage(ctx, message); err != nil {
		return err
	}

	arena := arenaPool.Get()
	defer arenaPool.Put(arena)

	ack := arena.NewObject()
	ack.Set("id", messageId)

	bytes := bytesPool.Get().([]byte)
	defer func() {
		bytes = bytes[:0]
		bytesPool.Put(bytes)
	}()

	// The library supports concurrent writes, we take a read lock to prevent race conditions
	// in case the connection is reestablished.
	s.RLock()
	err = s.conn.Write(ctx, websocket.MessageText, ack.MarshalTo(bytes))
	s.RUnlock()
	if err != nil {
		if err := s.reestablishConnection(ctx, err); err != nil {
			return err
		}
	}

	metrics.G.RecordMessagesAcked(ctx)

	return nil
}
