package graphmetrics

import (
	"time"

	"github.com/graphmetrics/logger-go"

	"github.com/graphmetrics/graphmetrics-go/internal"
)

const flushInterval = 1 * time.Minute

type Aggregator struct {
	metrics       *internal.UsageMetrics
	serverVersion string

	flushTicker   *time.Ticker
	fieldChan     chan *FieldMessage
	operationChan chan *OperationMessage
	stopChan      chan interface{}
	sender        *Sender

	logger logger.Logger
}

func NewAggregator(cfg *Configuration) *Aggregator {
	return &Aggregator{
		metrics:       internal.NewUsageMetrics(),
		serverVersion: cfg.ServerVersion,
		flushTicker:   time.NewTicker(flushInterval),
		fieldChan:     make(chan *FieldMessage, cfg.getFieldBufferSize()),
		operationChan: make(chan *OperationMessage, cfg.getOperationBufferSize()),
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
		case o := <-a.operationChan:
			a.processOperation(o)
		case f := <-a.fieldChan:
			a.processField(f)
			break
		}
	}
}

func (a *Aggregator) Stop() error {
	a.logger.Debug("stopping aggregator", nil)
	a.stopChan <- nil
	close(a.fieldChan)
	for msg := range a.fieldChan {
		a.processField(msg)
	}
	a.flush()
	return a.sender.Stop()
}

func (a *Aggregator) PushField(msg *FieldMessage) {
	if msg.TypeName[0:2] == "__" || msg.FieldName[0:2] == "__" {
		return
	}
	select {
	case a.fieldChan <- msg:
		return
	default:
		a.logger.Warn("graphmetrics aggregator field buffer overflowing, dropping message", nil)
	}
}

func (a *Aggregator) PushOperation(msg *OperationMessage) {
	select {
	case a.operationChan <- msg:
		return
	default:
		a.logger.Warn("graphmetrics aggregator operation buffer overflowing, dropping message", nil)
	}
}

func (a *Aggregator) processField(msg *FieldMessage) {
	// Find field metrics
	metrics := a.metrics.FindContextMetrics(msg.Client.Name, msg.Client.Version, a.serverVersion)
	typeMetrics := metrics.FindTypeMetrics(msg.TypeName)
	fieldMetrics := typeMetrics.FindFieldMetrics(msg.FieldName)

	// Insert message
	err := fieldMetrics.Histogram.Add(float64(msg.Duration))
	if err != nil {
		a.logger.Error("unable to insert field duration", map[string]interface{}{
			"error":    err,
			"duration": msg.Duration,
			"field":    msg.FieldName,
			"type":     msg.TypeName,
		})
		return
	}
	fieldMetrics.ErrorCount += internal.Bool2Int(msg.Error != nil)
	fieldMetrics.Count += 1
	fieldMetrics.ReturnType = msg.ReturnType
}

func (a *Aggregator) processOperation(msg *OperationMessage) {
	// Find operations metric
	metrics := a.metrics.FindContextMetrics(msg.Client.Name, msg.Client.Version, a.serverVersion)
	operationMetrics := metrics.FindOperationMetrics(msg.Hash)

	// Insert message
	err := operationMetrics.Histogram.Add(float64(msg.Duration))
	if err != nil {
		a.logger.Error("unable to insert operation duration", map[string]interface{}{
			"error":     err,
			"duration":  msg.Duration,
			"operation": msg.Name,
		})
		return
	}
	operationMetrics.ErrorCount += internal.Bool2Int(msg.HasErrors)
	operationMetrics.Count += 1
}

func (a *Aggregator) flush() {
	if len(a.metrics.Metrics) == 0 {
		return
	}
	metrics := a.metrics
	a.metrics = internal.NewUsageMetrics()
	metrics.Timestamp = time.Now() // We prefer end time as the TS
	a.sender.Send(metrics)
}
