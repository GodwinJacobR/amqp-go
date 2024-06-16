package amqppublisher

import (
	amqp_publisher "github.com/GodwinJacobR/amqp-go/internal/publisher"
)

type publisher interface {
	Publish(exchange string, payload interface{}) error
	Close() error
}

// Publisher represents a RabbitMQ publisher
type Publisher struct {
	publisher publisher
}

// NewPublisher creates a new Publisher instance
func NewPublisher(amqpURL, queueName string) (*Publisher, error) {

	p, err := amqp_publisher.NewPublisher(amqpURL, queueName)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		publisher: p,
	}, nil
}

// Publish sends a message to the RabbitMQ queue
func (p *Publisher) Publish(exchange string, payload interface{}) error {

	return p.publisher.Publish(exchange, payload)
	// TODO move this to internal
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {
	return p.publisher.Close()
}
