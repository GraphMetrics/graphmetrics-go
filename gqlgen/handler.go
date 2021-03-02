package graphmetricsgqlgen

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/graphmetrics/logger-go"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/graphmetrics/graphmetrics-go"
	"github.com/graphmetrics/graphmetrics-go/client"
	"github.com/graphmetrics/graphmetrics-go/signature"
)

type Extension interface {
	graphql.OperationInterceptor
	graphql.FieldInterceptor
	graphql.HandlerExtension

	Close() error
}

func NewExtension(cfg *graphmetrics.Configuration) Extension {
	agg := graphmetrics.NewAggregator(cfg)
	go agg.Start()
	return &extensionImpl{
		aggregator:      agg,
		clientExtractor: cfg.GetClientExtractor(),
		logger:          cfg.GetLogger(),
	}
}

type extensionImpl struct {
	aggregator      *graphmetrics.Aggregator
	clientExtractor client.Extractor
	schema          *ast.Schema

	logger logger.Logger
}

func (*extensionImpl) ExtensionName() string {
	return "GraphMetricsExtension"
}

func (e *extensionImpl) Validate(schema graphql.ExecutableSchema) error {
	e.schema = schema.Schema()
	return nil
}

func (e *extensionImpl) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	operation := graphql.GetOperationContext(ctx)
	caller := e.clientExtractor(ctx)
	sign, err := signature.OperationSignature(e.schema, operation.RawQuery, operation.OperationName)
	if err != nil {
		e.logger.Error("unable to build operation signature", map[string]interface{}{
			"err":       err,
			"operation": operation.OperationName,
		})
	}
	hash := signature.OperationHash(sign)

	handler := next(ctx)
	return func(ctx context.Context) *graphql.Response {
		res := handler(ctx)
		duration := time.Since(operation.Stats.OperationStart)
		e.aggregator.PushOperation(&graphmetrics.OperationMessage{
			Name:      operation.OperationName,
			Type:      string(operation.Operation.Operation),
			Hash:      hash,
			Signature: sign,
			HasErrors: len(res.Errors) > 0,
			Duration:  duration,
			Client:    caller,
		})

		return res
	}
}

func (e *extensionImpl) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	start := time.Now()

	field := graphql.GetFieldContext(ctx)
	caller := e.clientExtractor(ctx)
	res, err = next(ctx)
	duration := time.Since(start)
	e.aggregator.PushField(&graphmetrics.FieldMessage{
		TypeName:   field.Object,
		FieldName:  field.Field.Name,
		ReturnType: field.Field.Definition.Type.String(),
		Error:      err,
		Duration:   duration,
		Client:     caller,
	})

	return res, err
}

func (e *extensionImpl) Close() error {
	return e.aggregator.Stop()
}
