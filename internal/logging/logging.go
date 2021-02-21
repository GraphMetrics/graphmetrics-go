package logging

import (
	"github.com/graphmetrics/graphmetrics-go/internal/conversion"
	"github.com/graphmetrics/logger-go"
	"github.com/graphmetrics/logger-go/options"
	"github.com/hashicorp/go-retryablehttp"
)

type retryableLogger struct {
	parent logger.Logger
}

func NewRetryableLogger(parent logger.Logger) retryablehttp.LeveledLogger {
	return &retryableLogger{parent.WithOptions(options.CallerSkipOffset{Offset: 1})}
}

func (r retryableLogger) Debug(msg string, keysAndValues ...interface{}) {
	r.parent.Debug(msg, conversion.KeysAndValuesToMap(keysAndValues))
}

func (r retryableLogger) Info(msg string, keysAndValues ...interface{}) {
	r.parent.Info(msg, conversion.KeysAndValuesToMap(keysAndValues))
}

func (r retryableLogger) Warn(msg string, keysAndValues ...interface{}) {
	r.parent.Warn(msg, conversion.KeysAndValuesToMap(keysAndValues))
}

func (r retryableLogger) Error(msg string, keysAndValues ...interface{}) {
	r.parent.Error(msg, conversion.KeysAndValuesToMap(keysAndValues))
}
