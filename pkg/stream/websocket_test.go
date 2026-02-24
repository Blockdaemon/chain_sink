package stream

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/stream/mock_stream"
	"github.com/coder/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	testMessageOne = `{"type": "greeting", "message": "Hello, WebSocket!", "id":"test-message-one"}`
)

var (
	testServer     *TestServer
	testServerPort int
)

func TestMain(m *testing.M) {
	logger.Init(logger.Config{Level: "debug"})

	var err error
	testServer, testServerPort, err = NewTestServer(context.Background())
	if err != nil {
		logger.Log.Error("error creating test server", zap.Error(err))
		os.Exit(1)
	}

	exitCode := m.Run()
	_ = testServer.Close(context.Background())
	os.Exit(exitCode)
}

func TestWebsocket_noAckMode(t *testing.T) {
	const targetId = "e61f5784-8d00-4251-b809-c631c58840ae"

	ctx, cancel := context.WithCancel(context.Background())

	adapter := mock_stream.NewMockAdapter(t)
	adapter.EXPECT().HandleMessage(mock.Anything, mock.MatchedBy(func(message []byte) bool {
		if string(message) == testMessageOne {
			cancel()
			return true
		}
		return false
	})).Return(nil)

	stream, err := NewChainWatchStream(context.Background(), Config{
		URL:            fmt.Sprintf("ws://localhost:%d/targets/%s/websocket", testServerPort, targetId),
		Mode:           StreamModeNoAck,
		WorkerPoolSize: 1,
	})
	require.NoError(t, err)

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return stream.ForwardMessagesToAdapter(gCtx, adapter)
	})

	group.Go(func() error {
		serverConn, err := testServer.waitForConn(targetId, 5*time.Second)
		require.NoError(t, err)
		return serverConn.Conn.Write(gCtx, websocket.MessageText, []byte(testMessageOne))
	})

	err = group.Wait()
	assert.ErrorIs(t, err, context.Canceled)
}

func TestWebsocket_AckMode(t *testing.T) {
	const targetId = "921bc401-c875-425a-8448-81ea5f2e6f3d"

	ctx, cancel := context.WithCancel(context.Background())

	adapter := mock_stream.NewMockAdapter(t)
	adapter.EXPECT().HandleMessage(mock.Anything, []byte(testMessageOne)).Return(nil)

	stream, err := NewChainWatchStream(context.Background(), Config{
		URL:            fmt.Sprintf("ws://localhost:%d/targets/%s/websocket", testServerPort, targetId),
		Mode:           StreamModeAck,
		WorkerPoolSize: 1,
	})
	require.NoError(t, err)

	var testSuccess bool

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return stream.ForwardMessagesToAdapter(gCtx, adapter)
	})

	serverConn, err := testServer.waitForConn(targetId, 5*time.Second)
	require.NoError(t, err)

	group.Go(func() error {
		return serverConn.Conn.Write(gCtx, websocket.MessageText, []byte(testMessageOne))
	})

	group.Go(func() error {
		messageType, message, err := serverConn.Conn.Read(gCtx)
		assert.NoError(t, err)
		assert.Equal(t, websocket.MessageText, messageType)
		assert.JSONEq(t, `{"id":"test-message-one"}`, string(message))
		cancel()
		testSuccess = true
		return nil
	})

	err = group.Wait()
	assert.ErrorIs(t, err, context.Canceled)
	assert.True(t, testSuccess)
}

// Below code is a test server for the websocket connection. It is used to test the websocket connection in isolation.
type TestWebsocketConn struct {
	Conn *websocket.Conn
}

func NewTestWebsocketConn(conn *websocket.Conn) *TestWebsocketConn {
	return &TestWebsocketConn{
		Conn: conn,
	}
}

type TestServer struct {
	sync.RWMutex
	conns  map[string]*TestWebsocketConn
	server *echo.Echo
}

func NewTestServer(ctx context.Context) (*TestServer, int, error) {
	self := &TestServer{
		server: echo.New(),
		conns:  make(map[string]*TestWebsocketConn),
	}

	self.server.GET("/targets/:id/websocket", self.Handler)

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Log.Error("could not listen on a port", zap.Error(err))
		return nil, 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	go func() {
		if err := self.server.Start(fmt.Sprintf(":%d", port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("error starting server", zap.Error(err))
		}
	}()

	// Wait for server to be listening so tests don't race (CI is slower and often hits the race).
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err == nil {
			c.Close()
			return self, port, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, 0, fmt.Errorf("server did not become ready on port %d", port)
}

func (t *TestServer) Handler(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return fmt.Errorf("id is required")
	}

	conn, err := websocket.Accept(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	connWrapper := NewTestWebsocketConn(conn)
	t.Lock()
	if existing, ok := t.conns[id]; ok {
		_ = existing.Conn.Close(websocket.StatusNormalClosure, "replaced by new connection")
	}
	t.conns[id] = connWrapper
	t.Unlock()

	return nil
}

func (t *TestServer) GetConn(id string) (*TestWebsocketConn, error) {
	t.RLock()
	defer t.RUnlock()
	conn, ok := t.conns[id]
	if !ok {
		return nil, fmt.Errorf("connection not found")
	}
	return conn, nil
}

// waitForConn polls GetConn until the connection appears or timeout. Prevents races where
// the client has dialed but the server Handler has not yet stored the conn (common in CI).
func (t *TestServer) waitForConn(id string, timeout time.Duration) (*TestWebsocketConn, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := t.GetConn(id)
		if err == nil {
			return conn, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return t.GetConn(id)
}

func (t *TestServer) Close(ctx context.Context) error {
	t.Lock()
	defer t.Unlock()
	for _, conn := range t.conns {
		_ = conn.Conn.Close(websocket.StatusNormalClosure, "test server closing")
	}
	t.conns = make(map[string]*TestWebsocketConn)
	return t.server.Shutdown(ctx)
}
