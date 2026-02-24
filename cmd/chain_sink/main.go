package main

import (
	"context"
	"fmt"

	"github.com/blockdaemon/chain_sink/pkg/appctx"
	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/blockdaemon/chain_sink/pkg/metrics"
	"github.com/blockdaemon/chain_sink/pkg/stream"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {

	cfg, err := LoadConfig()
	if err != nil {
		logger.Init(logger.Config{Level: "error"})
		logger.Log.Fatal("error loading config", zap.Error(err))
	}

	logger.Init(cfg.Logger)

	ctx, cancel := appctx.Context()
	defer cancel()

	adapter, err := buildAdapter(ctx, cfg.Adapter)
	if err != nil {
		logger.Log.Fatal("error building adapter", zap.Error(err))
	}

	group, gCtx := errgroup.WithContext(ctx)

	for range cfg.StreamCount {
		group.Go(func() error {
			chainWatchStream, err := stream.NewChainWatchStream(gCtx, cfg.Stream)
			if err != nil {
				return err
			}
			return chainWatchStream.ForwardMessagesToAdapter(gCtx, adapter)
		})
	}

	if cfg.Metrics.Enabled {
		_, err := metrics.Init(gCtx)
		if err != nil {
			logger.Log.Fatal("error initializing metrics", zap.Error(err))
		}

		server := echo.New()
		server.GET("/metrics", echo.WrapHandler(metrics.Handler()))
		group.Go(func() error {
			logger.Log.Info("starting metrics server", zap.Int("port", cfg.Metrics.Port))
			return server.Start(fmt.Sprintf(":%d", cfg.Metrics.Port))
		})

		group.Go(func() error {
			<-gCtx.Done()
			_ = server.Shutdown(context.Background())
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		logger.Log.Fatal("error running streams", zap.Error(err))
	}
}
