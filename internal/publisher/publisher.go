package amqppublisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

	logger    internal.Logger
	tracer    trace.Tracer
	confirmCh *chan amqp.Confirmation
}

func NewPublisher(amqpUrl, queueName string, logger internal.Logger, config PublisherConfig) (*Publisher, error) {
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

	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &Publisher{
		conn:      conn,
		ch:        ch,
		queue:     queue,
		logger:    logger,
		tracer:    config.tracer,
		confirmCh: &confirmCh,
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
		handlePublishError(err, span, exchange, event.EventName, nil)
		return err
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
		handlePublishError(err, span, exchange, event.EventName, nil)
		return err
	}

	select {
	case confirm := <-*p.confirmCh:
		if !confirm.Ack {
			handlePublishError(errors.New("message not acknowledged by RabbitMQ"), span, exchange, event.EventName, nil)
			return err
		}
		handlePublishSuccess(exchange, event.EventName, nil)
	case <-time.After(5 * time.Second): // Timeout after 5 seconds
		handlePublishError(errors.New("timeout waiting for publisher confirmation"), span, exchange, event.EventName, nil)
		return err
	}
	return nil
}

func (p *Publisher) PublishWithNotificationBack(ctx context.Context, exchange string, event go_amqp.Event, notifyCh chan<- error) error {
	if notifyCh == nil {
		return fmt.Errorf("notifyCh is nil")
	}

	go func() {

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
			handlePublishError(err, span, exchange, event.EventName, notifyCh)
			return
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
			handlePublishError(err, span, exchange, event.EventName, notifyCh)
			return
		}

		select {
		case confirm := <-*p.confirmCh:
			if !confirm.Ack {
				handlePublishError(errors.New("message not acknowledged by RabbitMQ"), span, exchange, event.EventName, notifyCh)
				return
			}
			handlePublishSuccess(exchange, event.EventName, notifyCh)
		case <-time.After(5 * time.Second): // Timeout after 5 seconds
			handlePublishError(errors.New("timeout waiting for publisher confirmation"), span, exchange, event.EventName, notifyCh)
		}
	}()
	return nil
}

// Close cleans up the RabbitMQ connection and channel
func (p *Publisher) Close() error {

	return errors.Join(
		p.ch.Close(),
		p.conn.Close(),
	)
}

func handlePublishError(err error, span trace.Span, exchange, eventName string, notifyCh chan<- error) {
	span.RecordError(err)
	go_amqp.EventPublishFailed(exchange, eventName)
	if notifyCh == nil {
		notifyCh <- err
	}
}

func handlePublishSuccess(exchange, eventName string, notifyCh chan<- error) {
	go_amqp.EventPublishSucceeded(exchange, eventName)
	if notifyCh == nil {
		notifyCh <- nil
	}
}
