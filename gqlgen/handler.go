package graphmetricsgqlgen

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/graphmetrics/graphmetrics-go"
	"github.com/graphmetrics/graphmetrics-go/client"
)

type Extension interface {
	graphql.OperationInterceptor
	graphql.FieldInterceptor
	graphql.HandlerExtension

	Close() error
}

func NewExtension(cfg *graphmetrics.Configuration) Extension {
	clientExt := cfg.ClientExtractor
	if clientExt == nil {
		clientExt = func(context.Context) client.Details {
			return client.Details{}
		}
	}
	agg := graphmetrics.NewAggregator(cfg)
	go agg.Start()
	return &extensionImpl{
		agg:       agg,
		clientExt: clientExt,
	}
}

type extensionImpl struct {
	agg       *graphmetrics.Aggregator
	clientExt client.Extractor
}

func (*extensionImpl) ExtensionName() string {
	return "GraphMetricsExtension"
}

func (*extensionImpl) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (e *extensionImpl) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// NOT IMPLEMENTED YET
	return next(ctx)
}

func (e *extensionImpl) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	field := graphql.GetFieldContext(ctx)
	start := time.Now()
	res, err = next(ctx)
	duration := time.Since(start)
	e.agg.PushField(&graphmetrics.FieldMessage{
		TypeName:   field.Object,
		FieldName:  field.Field.Name,
		ReturnType: field.Field.Definition.Type.NamedType,
		Error:      err,
		Duration:   duration,
		Client:     e.clientExt(ctx),
	})
	return res, err
}

func (e *extensionImpl) Close() error {
	e.agg.Stop()
	return nil
}
