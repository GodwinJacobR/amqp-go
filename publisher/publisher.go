package amqppublisher

import (
	"context"

	amqp_go "github.com/GodwinJacobR/go-amqp"
	"github.com/GodwinJacobR/go-amqp/internal"
	amqp_publisher "github.com/GodwinJacobR/go-amqp/internal/publisher"
	"go.opentelemetry.io/otel"
)

type publisher interface {
	Publish(ctx context.Context, exchange string, payload amqp_go.Event) error
	Close() error
}

// Publisher represents a RabbitMQ publisher
type Publisher struct {
	inner publisher
}

// NewPublisher creates a new Publisher instance
func NewPublisher(amqpURL string, logger internal.Logger, queueName string, options ...Option) (*Publisher, error) {
	config := publisherConfig{
		tracer: otel.Tracer("amqp-go"),
	}
	for _, o := range options {
		o(&config)
	}
	p, err := amqp_publisher.NewPublisher(amqpURL, queueName, logger, config.tracer)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		inner: p,
	}, nil
}

// Publish sends a message to the RabbitMQ queue
func (p *Publisher) Publish(ctx context.Context, exchange, eventName string, payload any, opts ...amqp_go.EventOption) error {
	event := amqp_go.NewEvent(ctx, payload, eventName, opts...)

	return p.inner.Publish(ctx, exchange, event)
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {
	return p.inner.Close()
}
