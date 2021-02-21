package graphmetrics

import (
	"context"
	"time"

	"github.com/graphmetrics/logger-go"

	"github.com/graphmetrics/graphmetrics-go/client"
)

const (
	defaultEndpoint            = "api.graphmetrics.io"
	defaultFieldBufferSize     = 1000
	defaultOperationBufferSize = 20
	defaultStopTimeout         = 10 * time.Second
)

type Configuration struct {
	ApiKey          string
	ServerVersion   string
	ClientExtractor client.Extractor
	Logger          logger.Logger
	Advanced        *AdvancedConfiguration
}

type AdvancedConfiguration struct {
	FieldBufferSize     int // If field metrics are dropped consider increasing it
	OperationBufferSize int // If operation metrics are dropped consider increasing it
	Endpoint            string
	Http                bool
	Debug               bool
	StopTimeout         time.Duration
}

func (c *Configuration) getEndpoint() string {
	if c.Advanced != nil && c.Advanced.Endpoint != "" {
		return c.Advanced.Endpoint
	}
	return defaultEndpoint
}

func (c *Configuration) getProtocol() string {
	if c.Advanced != nil && c.Advanced.Http {
		return "http"
	}
	return "https"
}

func (c *Configuration) getFieldBufferSize() int {
	if c.Advanced != nil && c.Advanced.FieldBufferSize != 0 {
		return c.Advanced.FieldBufferSize
	}
	return defaultFieldBufferSize
}

func (c *Configuration) getOperationBufferSize() int {
	if c.Advanced != nil && c.Advanced.OperationBufferSize != 0 {
		return c.Advanced.OperationBufferSize
	}
	return defaultOperationBufferSize
}

func (c *Configuration) getDebug() bool {
	if c.Advanced != nil {
		return c.Advanced.Debug
	}
	return false
}

func (c *Configuration) getStopTimeout() time.Duration {
	if c.Advanced != nil && c.Advanced.StopTimeout != 0 {
		return c.Advanced.StopTimeout
	}
	return defaultStopTimeout
}

func (c *Configuration) getLogger() logger.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return logger.NewDefault(c.getDebug())
}

func (c *Configuration) getClientExtractor() client.Extractor {
	if c.ClientExtractor != nil {
		return c.ClientExtractor
	}
	return func(context.Context) client.Details {
		return client.Details{}
	}
}
