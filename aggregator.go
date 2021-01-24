package graphmetrics

import (
	"time"

	"github.com/graphmetrics/graphmetrics-go/internal"
)

const flushInterval = 1 * time.Minute

type Aggregator struct {
	metrics       *internal.UsageMetrics
	serverVersion string

	flushTicker *time.Ticker
	fieldChan   chan *FieldMessage
	stopChan    chan interface{}
	sender      *Sender

	logger Logger
}

func NewAggregator(cfg *Configuration) *Aggregator {
	return &Aggregator{
		metrics:       internal.NewUsageMetrics(),
		serverVersion: cfg.ServerVersion,
		flushTicker:   time.NewTicker(flushInterval),
		fieldChan:     make(chan *FieldMessage, cfg.getFieldBufferSize()),
		stopChan:      make(chan interface{}),
		sender:        NewSender(cfg),
		logger:        cfg.getLogger(),
	}
}

func (a *Aggregator) Start() {
	go a.sender.Start()
	for {
		select {
		case <-a.stopChan:
			return
		case <-a.flushTicker.C:
			a.flush()
		case f := <-a.fieldChan:
			a.processField(f)
		}
	}
}

func (a *Aggregator) Stop() {
	a.stopChan <- nil
	for msg := range a.fieldChan {
		a.processField(msg)
	}
	a.sender.Stop()
}

func (a *Aggregator) PushField(msg *FieldMessage) {
	if msg.TypeName[0:2] == "__" || msg.FieldName[0:2] == "__" {
		return
	}
	select {
	case a.fieldChan <- msg:
		return
	default:
		a.logger.Warn("graphmetrics aggregator buffer overflowing, dropping message", nil)
	}
}

func (a *Aggregator) processField(msg *FieldMessage) {
	// Find field metric
	typesMetrics := a.metrics.FindTypesMetrics(msg.Client.Name, msg.Client.Version, a.serverVersion)
	typeMetric := typesMetrics.FindTypeMetric(msg.TypeName)
	fieldMetric := typeMetric.FindFieldMetric(msg.FieldName)

	// Insert message
	err := fieldMetric.Histogram.Add(float64(msg.Duration))
	if err != nil {
		a.logger.Error("unable to insert field duration", map[string]interface{}{
			"error":    err,
			"duration": msg.Duration,
			"field":    msg.FieldName,
			"type":     msg.TypeName,
		})
		return
	}
	fieldMetric.ErrorCount += internal.Bool2Int(msg.Error != nil)
	fieldMetric.ErrorCount += 1
	fieldMetric.ReturnType = msg.ReturnType
}

func (a *Aggregator) flush() {
	if len(a.metrics.Types) == 0 {
		return
	}
	metrics := a.metrics
	a.metrics = internal.NewUsageMetrics()
	metrics.Timestamp = time.Now() // We prefer end time as the TS
	a.sender.Send(metrics)
}
