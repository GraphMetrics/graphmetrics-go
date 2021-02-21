package graphmetrics

import (
	"time"

	"github.com/graphmetrics/graphmetrics-go/client"
)

type FieldMessage struct {
	TypeName   string
	FieldName  string
	ReturnType string
	Error      error
	Duration   time.Duration
	Client     client.Details
}

type OperationMessage struct {
	Name      string
	Hash      string
	HasErrors bool
	Duration  time.Duration
	Client    client.Details
}
