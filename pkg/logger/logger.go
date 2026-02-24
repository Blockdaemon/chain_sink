package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log = zap.NewNop()
)

// Init creates a new global logger, if not called the global logger will be a noop implementation.
func Init(cfg Config) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = getLogLevel(cfg.Level)
	zapConfig.DisableStacktrace = cfg.DisableStacktrace
	zapConfig.Development = cfg.Development

	var err error
	Log, err = zapConfig.Build()
	if err != nil {
		fmt.Println("failed to init the logger")
		panic(err)
	}
}

func getLogLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "error":
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zapcore.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}
