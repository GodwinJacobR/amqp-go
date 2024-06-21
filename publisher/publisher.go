package amqppublisher

import (
	"context"

	go_amqp "github.com/GodwinJacobR/go-amqp"
	"github.com/GodwinJacobR/go-amqp/internal"
	amqp_publisher "github.com/GodwinJacobR/go-amqp/internal/publisher"
)

type publisher interface {
	Publish(ctx context.Context, exchange string, payload go_amqp.Event) error
	PublishWithNotificationBack(ctx context.Context, exchange string, payload go_amqp.Event, notifyCh chan<- error) error
	Close() error
}

type PublisherConfig = amqp_publisher.PublisherConfig
type Option = amqp_publisher.Option

// Publisher represents a RabbitMQ publisher
type Publisher struct {
	inner publisher
}

// NewPublisher creates a new Publisher instance
func NewPublisher(amqpURL string, logger internal.Logger, queueName string, options ...Option) (*Publisher, error) {
	config := amqp_publisher.DefaultConfigs()
	for _, o := range options {
		o(&config)
	}
	p, err := amqp_publisher.NewPublisher(amqpURL, queueName, logger, config)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		inner: p,
	}, nil
}

// Publish sends a message to the RabbitMQ queue
func (p *Publisher) Publish(ctx context.Context, exchange, eventName string, payload any, opts ...go_amqp.EventOption) error {
	event := go_amqp.NewEvent(ctx, payload, eventName, opts...)

	return p.inner.Publish(ctx, exchange, event)
}

func (p *Publisher) PublishWithNotificationBack(ctx context.Context, exchange, eventName string, payload any, notifyCh chan<- error, opts ...go_amqp.EventOption) error {
	event := go_amqp.NewEvent(ctx, payload, eventName, opts...)

	return p.inner.PublishWithNotificationBack(ctx, exchange, event, notifyCh)
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {
	return p.inner.Close()
}
