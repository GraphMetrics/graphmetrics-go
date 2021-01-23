package internal

import (
	"encoding/json"

	"github.com/graphmetrics/ddsketch-go/ddsketch"
)

const (
	clientsAllocation = 1
	typesAllocation   = 10
	fieldsAllocation  = 10
	relativeAccuracy  = 0.01
)

type Histogram struct {
	Indexes []int16 `json:"indexes"`
	Counts  []int32 `json:"counts"`
}

type FieldMetric struct {
	ReturnType string             `json:"returnType"`
	Count      int32              `json:"count"`
	ErrorCount int32              `json:"errorCount"`
	Histogram  *ddsketch.DDSketch `json:"-"`
}

func (f *FieldMetric) MarshalJSON() ([]byte, error) {
	// extract keys and counts from histogram bins
	// conservative size of half the count will be in the same bin
	indexes := make([]int16, 0, f.Count/2)
	counts := make([]int32, 0, f.Count/2)
	for b := range f.Histogram.Bins() {
		// we made some guarantees in the sketch so all indexes are within 16 bits
		indexes = append(indexes, int16(b.Index()))
		counts = append(counts, b.Count())
	}

	type Alias FieldMetric
	return json.Marshal(&struct {
		Histogram Histogram
		*Alias
	}{
		Histogram: Histogram{Indexes: indexes, Counts: counts},
		Alias:     (*Alias)(f),
	})
}

type TypeMetric struct {
	Fields map[string]*FieldMetric `json:"fields"`
}

func (t *TypeMetric) FindFieldMetric(fieldName string) *FieldMetric {
	if v, ok := t.Fields[fieldName]; ok {
		return v
	} else {
		h, _ := ddsketch.LogUnboundedDenseDDSketch(relativeAccuracy)
		t.Fields[fieldName] = &FieldMetric{
			Histogram: h,
		}
		return t.Fields[fieldName]
	}
}

type MetricsContext struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	ServerVersion string `json:"serverVersion"`
}

type ContextualizedTypesMetrics struct {
	Context MetricsContext         `json:"context"`
	Types   map[string]*TypeMetric `json:"types"`
}

func (t *ContextualizedTypesMetrics) FindTypeMetric(typeName string) *TypeMetric {
	if v, ok := t.Types[typeName]; ok {
		return v
	} else {
		t.Types[typeName] = &TypeMetric{
			Fields: make(map[string]*FieldMetric, fieldsAllocation),
		}
		return t.Types[typeName]
	}
}

type UsageMetrics struct {
	Types []ContextualizedTypesMetrics `json:"types"`
}

func (u *UsageMetrics) FindTypesMetrics(ClientName string, ClientVersion string, ServerVersion string) *ContextualizedTypesMetrics {
	for _, t := range u.Types {
		if t.Context.ClientName == ClientName &&
			t.Context.ClientVersion == ClientVersion &&
			t.Context.ServerVersion == ServerVersion {
			return &t
		}
	}
	t := ContextualizedTypesMetrics{
		Context: MetricsContext{
			ClientName:    ClientName,
			ClientVersion: ClientVersion,
			ServerVersion: ServerVersion,
		},
		Types: make(map[string]*TypeMetric, typesAllocation),
	}
	u.Types = append(u.Types, t)
	return &t
}

func NewUsageMetrics() *UsageMetrics {
	return &UsageMetrics{
		Types: make([]ContextualizedTypesMetrics, 0, clientsAllocation),
	}
}
