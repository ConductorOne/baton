package logging

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogFormatJSON    = "json"
	LogFormatConsole = "console"
)

type Option func(*zap.Config)

func WithLogLevel(level string) Option {
	return func(c *zap.Config) {
		ll := zapcore.DebugLevel
		_ = ll.Set(level)
		c.Level.SetLevel(ll)
	}
}

func WithLogFormat(format string) Option {
	return func(c *zap.Config) {
		switch format {
		case LogFormatJSON:
			c.Encoding = LogFormatJSON
		case LogFormatConsole:
			c.Encoding = LogFormatConsole
		default:
			c.Encoding = LogFormatJSON
		}
	}
}

func WithOutputPaths(paths []string) Option {
	return func(c *zap.Config) {
		c.OutputPaths = paths
	}
}

// Init creates a new zap logger and attaches it to the provided context.
func Init(ctx context.Context, opts ...Option) (context.Context, error) {
	zc := zap.NewProductionConfig()
	zc.Sampling = nil
	zc.DisableStacktrace = true

	for _, opt := range opts {
		opt(&zc)
	}

	l, err := zc.Build()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(l)

	l.Debug("Logger created!", zap.String("log_level", zc.Level.String()))

	return ctxzap.ToContext(ctx, l), nil
}
