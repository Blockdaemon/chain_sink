// Package appctx provides a context that can be cancelled using interrupts.
package appctx

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/blockdaemon/chain_sink/pkg/logger"
)

var once sync.Once
var ctx context.Context
var cancel context.CancelFunc

// Context returns the application context that closes when the app gets an interrupt.
// It is safe to call this function multiple times, it will return the same context object.
func Context() (context.Context, context.CancelFunc) {
	once.Do(func() {
		ctx, cancel = context.WithCancel(context.Background())
		go func() {
			defer cancel()
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			logger.Log.Info("cancelling application context")
		}()
	})
	return ctx, cancel
}
