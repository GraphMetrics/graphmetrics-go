package graphmetrics

import (
	"context"

	"github.com/graphmetrics/graphmetrics-go/client"
)

const (
	defaultEndpoint        = "api.graphmetrics.io"
	defaultFieldBufferSize = 1000
)

type Configuration struct {
	ApiKey          string
	ServerVersion   string
	ClientExtractor client.Extractor
	Logger          Logger
	Advanced        *AdvancedConfiguration
}

type AdvancedConfiguration struct {
	FieldBufferSize int // If field metrics are dropped consider increasing it
	Endpoint        string
}

func (c *Configuration) getEndpoint() string {
	if c.Advanced != nil && c.Advanced.Endpoint != "" {
		return c.Advanced.Endpoint
	}
	return defaultEndpoint
}

func (c *Configuration) getFieldBufferSize() int {
	if c.Advanced != nil && c.Advanced.FieldBufferSize != 0 {
		return c.Advanced.FieldBufferSize
	}
	return defaultFieldBufferSize
}

func (c *Configuration) getLogger() Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return &defaultLogger{}
}

func (c *Configuration) getClientExtractor() client.Extractor {
	if c.ClientExtractor != nil {
		return c.ClientExtractor
	}
	return func(context.Context) client.Details {
		return client.Details{}
	}
}
