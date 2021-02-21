package models

import "time"

type OperationDefinition struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
}

type UsageDefinitions struct {
	Timestamp  time.Time             `json:"timestamp"`
	Operations []OperationDefinition `json:"operations"`
}

func NewUsageDefinitions() *UsageDefinitions {
	return &UsageDefinitions{
		Timestamp:  time.Time{},
		Operations: make([]OperationDefinition, 0, 5),
	}
}
