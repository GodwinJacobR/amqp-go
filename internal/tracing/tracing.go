package tracing

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// inject the span context to amqp table
func InjectToHeaders(ctx context.Context) amqp.Table {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	header := amqp.Table{}
	for k, v := range carrier {
		header[k] = v
	}
	return header
}
