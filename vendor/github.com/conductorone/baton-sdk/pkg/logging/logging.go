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

// Init creates a new zap logger and attaches it to the provided context.
func Init(ctx context.Context, format string, level string) (context.Context, error) {
	zc := zap.NewProductionConfig()
	zc.Sampling = nil
	zc.DisableStacktrace = true

	ll := zapcore.DebugLevel
	err := ll.Set(level)
	if err != nil {
		return ctx, err
	}
	zc.Level.SetLevel(ll)

	zc.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	switch format {
	case LogFormatJSON:
		zc.Encoding = LogFormatJSON
	case LogFormatConsole:
		zc.Encoding = LogFormatConsole
	default:
		zc.Encoding = LogFormatJSON
	}

	l, err := zc.Build()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(l)

	l.Debug("Logger created!", zap.String("log_level", zc.Level.String()))

	return ctxzap.ToContext(ctx, l), nil
}
