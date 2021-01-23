package graphmetrics

import "github.com/graphmetrics/graphmetrics-go/client"

type Configuration struct {
	ApiKey          string
	ServerVersion   string
	ClientExtractor client.Extractor
	Endpoint        string
	Logger          Logger
}
