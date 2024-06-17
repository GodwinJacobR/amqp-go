package amqppublisher

import (
	"context"
	"fmt"

	amqp_go "github.com/GodwinJacobR/amqp-go"
	"github.com/GodwinJacobR/amqp-go/internal"
	amqp_publisher "github.com/GodwinJacobR/amqp-go/internal/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type publisher interface {
	Publish(ctx context.Context, exchange string, payload amqp_go.Event) error
	Close() error
}

// Publisher represents a RabbitMQ publisher
type Publisher struct {
	inner  publisher
	logger internal.Logger
	tracer trace.Tracer
}

// NewPublisher creates a new Publisher instance
func NewPublisher(amqpURL string, logger internal.Logger, queueName string, options ...Option) (*Publisher, error) {
	config := publisherConfig{
		tracer: otel.Tracer("amqp-go"),
	}
	for _, o := range options {
		o(&config)
	}
	p, err := amqp_publisher.NewPublisher(amqpURL, queueName)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		inner:  p,
		logger: logger,
		tracer: config.tracer,
	}, nil
}

// Publish sends a message to the RabbitMQ queue
func (p *Publisher) Publish(ctx context.Context, exchange, eventName string, payload any, opts ...amqp_go.EventOption) error {
	ctx, span := p.tracer.Start(
		ctx,
		fmt.Sprintf("%s %s", exchange, "PublishEvent"),
		trace.WithAttributes(
			attribute.String("event_name", eventName),
		),
	)
	defer span.End()

	event := amqp_go.NewEvent(ctx, payload, eventName, opts...)

	err := p.inner.Publish(ctx, exchange, event)
	if err != nil {
		span.RecordError(err)
		amqp_go.EventPublishFailed(exchange, eventName)
		return err
	}

	amqp_go.EventPublishSucceeded(exchange, eventName)
	return nil
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {
	return p.inner.Close()
}
