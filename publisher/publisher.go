package amqppublisher

import (
	"context"

	amqp_go "github.com/GodwinJacobR/amqp-go"
	amqp_publisher "github.com/GodwinJacobR/amqp-go/internal/publisher"
)

type publisher interface {
	Publish(exchange string, payload amqp_go.Event) error
	Close() error
}

// Publisher represents a RabbitMQ publisher
type Publisher struct {
	inner publisher
}

// NewPublisher creates a new Publisher instance
func NewPublisher(amqpURL, queueName string) (*Publisher, error) {

	p, err := amqp_publisher.NewPublisher(amqpURL, queueName)
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

	return p.inner.Publish(exchange, event)
	// TODO move this to internal
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {
	return p.inner.Close()
}
