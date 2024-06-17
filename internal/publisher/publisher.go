package amqp_publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	go_amqp "github.com/GodwinJacobR/go-amqp"
	"github.com/GodwinJacobR/go-amqp/internal"
	"github.com/GodwinJacobR/go-amqp/internal/tracing"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Publisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue

	logger internal.Logger
	tracer trace.Tracer
}

func NewPublisher(amqpUrl, queueName string, logger internal.Logger, tracer trace.Tracer) (*Publisher, error) {
	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	queue, err := ch.QueueDeclare(
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

	return &Publisher{
		conn:   conn,
		ch:     ch,
		queue:  queue,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, exchange string, event go_amqp.Event) error {
	ctx, span := p.tracer.Start(
		ctx,
		fmt.Sprintf("%s %s", exchange, "PublishEvent"),
		trace.WithAttributes(
			attribute.String("event_name", event.EventName),
		),
	)
	defer span.End()

	body, err := json.Marshal(event)
	if err != nil {
		handlePublishError(err, span, exchange, event.EventName)
		return fmt.Errorf("marshal message, err; %w", err)
	}

	headers := tracing.InjectToHeaders(ctx)
	headers["correlation_id"] = event.Metadata.CorrelationID

	err = p.ch.PublishWithContext(
		ctx,
		exchange,
		p.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Headers:       headers,
			ContentType:   "application/json",
			CorrelationId: event.Metadata.CorrelationID,
			MessageId:     event.Metadata.ID,
			Timestamp:     event.Metadata.Timestamp,
			Body:          body,
		})
	if err != nil {
		handlePublishError(err, span, exchange, event.EventName)
		return err
	}

	go_amqp.EventPublishSucceeded(exchange, event.EventName)
	return nil
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {

	return errors.Join(
		p.ch.Close(),
		p.conn.Close(),
	)
}

func handlePublishError(err error, span trace.Span, exchange, eventName string) {
	span.RecordError(err)
	go_amqp.EventPublishFailed(exchange, eventName)

}
