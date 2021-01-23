package client

import "context"

type Details struct {
	Name    string
	Version string
}

type Extractor func(context.Context) Details
