package models

import (
	"encoding/json"
	"time"

	"github.com/graphmetrics/sketches-go/ddsketch"
)

const (
	clientsAllocation    = 1
	typesAllocation      = 10
	operationsAllocation = 10
	fieldsAllocation     = 10
	relativeAccuracy     = 0.01
)

type Histogram struct {
	Indexes []int16 `json:"indexes"`
	Counts  []int32 `json:"counts"`
}

type FieldMetrics struct {
	ReturnType string             `json:"returnType"`
	Count      int32              `json:"count"`
	ErrorCount int32              `json:"errorCount"`
	Histogram  *ddsketch.DDSketch `json:"-"`
}

func (f *FieldMetrics) MarshalJSON() ([]byte, error) {
	// extract keys and counts from histogram bins
	// conservative size of half the count will be in the same bin
	indexes := make([]int16, 0, f.Count/2)
	counts := make([]int32, 0, f.Count/2)
	for b := range f.Histogram.Bins() {
		// we made some guarantees in the sketch so all indexes are within 16 bits
		indexes = append(indexes, int16(b.Index()))
		counts = append(counts, b.Count())
	}

	type Alias FieldMetrics
	return json.Marshal(&struct {
		Histogram Histogram
		*Alias
	}{
		Histogram: Histogram{Indexes: indexes, Counts: counts},
		Alias:     (*Alias)(f),
	})
}

type TypeMetrics struct {
	Fields map[string]*FieldMetrics `json:"fields"`
}

func (t *TypeMetrics) FindFieldMetrics(fieldName string) *FieldMetrics {
	if v, ok := t.Fields[fieldName]; ok {
		return v
	} else {
		h, _ := ddsketch.LogUnboundedDenseDDSketch(relativeAccuracy)
		t.Fields[fieldName] = &FieldMetrics{
			Histogram: h,
		}
		return t.Fields[fieldName]
	}
}

type OperationMetrics struct {
	Count      int32              `json:"count"`
	ErrorCount int32              `json:"errorCount"`
	Histogram  *ddsketch.DDSketch `json:"-"`
}

func (f *OperationMetrics) MarshalJSON() ([]byte, error) {
	// extract keys and counts from histogram bins
	// conservative size of half the count will be in the same bin
	indexes := make([]int16, 0, f.Count/2)
	counts := make([]int32, 0, f.Count/2)
	for b := range f.Histogram.Bins() {
		// we made some guarantees in the sketch so all indexes are within 16 bits
		indexes = append(indexes, int16(b.Index()))
		counts = append(counts, b.Count())
	}

	type Alias OperationMetrics
	return json.Marshal(&struct {
		Histogram Histogram
		*Alias
	}{
		Histogram: Histogram{Indexes: indexes, Counts: counts},
		Alias:     (*Alias)(f),
	})
}

type MetricsContext struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	ServerVersion string `json:"serverVersion"`
}

type ContextualizedUsageMetrics struct {
	Context    MetricsContext               `json:"context"`
	Types      map[string]*TypeMetrics      `json:"types"`
	Operations map[string]*OperationMetrics `json:"operations"`
}

func (t *ContextualizedUsageMetrics) FindTypeMetrics(typeName string) *TypeMetrics {
	if v, ok := t.Types[typeName]; ok {
		return v
	} else {
		t.Types[typeName] = &TypeMetrics{
			Fields: make(map[string]*FieldMetrics, fieldsAllocation),
		}
		return t.Types[typeName]
	}
}

func (t *ContextualizedUsageMetrics) FindOperationMetrics(operationHash string) *OperationMetrics {
	if v, ok := t.Operations[operationHash]; ok {
		return v
	} else {
		h, _ := ddsketch.LogUnboundedDenseDDSketch(relativeAccuracy)
		t.Operations[operationHash] = &OperationMetrics{
			Histogram: h,
		}
		return t.Operations[operationHash]
	}
}

type UsageMetrics struct {
	Timestamp time.Time                    `json:"timestamp"`
	Metrics   []ContextualizedUsageMetrics `json:"metrics"`
}

func (u *UsageMetrics) FindContextMetrics(ClientName string, ClientVersion string, ServerVersion string) *ContextualizedUsageMetrics {
	for _, t := range u.Metrics {
		if t.Context.ClientName == ClientName &&
			t.Context.ClientVersion == ClientVersion &&
			t.Context.ServerVersion == ServerVersion {
			return &t
		}
	}
	t := ContextualizedUsageMetrics{
		Context: MetricsContext{
			ClientName:    ClientName,
			ClientVersion: ClientVersion,
			ServerVersion: ServerVersion,
		},
		Types:      make(map[string]*TypeMetrics, typesAllocation),
		Operations: make(map[string]*OperationMetrics, operationsAllocation),
	}
	u.Metrics = append(u.Metrics, t)
	return &t
}

func NewUsageMetrics() *UsageMetrics {
	return &UsageMetrics{
		Timestamp: time.Time{},
		Metrics:   make([]ContextualizedUsageMetrics, 0, clientsAllocation),
	}
}
