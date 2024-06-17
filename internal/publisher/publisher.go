package amqp_publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	amqp_go "github.com/GodwinJacobR/amqp-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func NewPublisher(amqpURL, queueName string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{conn: conn, ch: ch, q: q}, nil
}

func (p *Publisher) Publish(ctx context.Context, exchange string, event amqp_go.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal message, err; %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		exchange,
		p.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			Headers:       amqp.Table{"correlation_id": event.Metadata.CorrelationID},
			ContentType:   "application/json",
			CorrelationId: event.Metadata.CorrelationID,
			MessageId:     event.Metadata.ID,
			Timestamp:     event.Metadata.Timestamp,
			Body:          body,
		})
	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s", body)
	return nil
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {

	return errors.Join(
		p.ch.Close(),
		p.conn.Close(),
	)
}
