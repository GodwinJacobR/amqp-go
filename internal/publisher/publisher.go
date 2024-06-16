package amqp_publisher

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/streadway/amqp"
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

func (p *Publisher) Publish(exchange string, payload interface{}) error {

	// TODO move this to internal
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message, err; %w", err)
	}

	err = p.ch.Publish(
		exchange,
		p.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
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
