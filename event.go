package amqp_go

import (
	"context"
	"time"

	"github.com/GodwinJacobR/amqp-go/correlation"
	"github.com/google/uuid"
)

type Event struct {
	Metadata `json:"metadata"`
	Payload  any `json:"payload"`
}

type Metadata struct {
	ID            string    `json:"id"`
	EventName     string    `json:"eventName"`
	CorrelationID string    `json:"correlationId"`
	Timestamp     time.Time `json:"timestamp"`
}

type eventOptions struct {
	eventID       string
	correlationID string
	timestamp     time.Time
}

var evtOpts eventOptions = eventOptions{
	eventID:       uuid.NewString(),
	correlationID: uuid.NewString(),
	timestamp:     time.Now(),
}

type EventOption func(*eventOptions)

func WithTimestamp(timestamp time.Time) EventOption {
	return func(e *eventOptions) {
		e.timestamp = timestamp
	}
}

func WithCorrelationID(correlationID string) EventOption {
	return func(e *eventOptions) {
		e.correlationID = correlationID
	}
}

func WithEventID(eventID string) EventOption {
	return func(e *eventOptions) {
		e.eventID = eventID
	}
}

func NewEvent(ctx context.Context, payload any, eventName string, opts ...EventOption) Event {
	eventOptions := eventOptions{
		correlationID: correlation.GetCorrelationID(ctx),
	}
	for _, opt := range opts {
		opt(&eventOptions)
	}

	return Event{
		Metadata: Metadata{
			ID:            evtOpts.eventID,
			EventName:     eventName,
			CorrelationID: evtOpts.correlationID,
			Timestamp:     evtOpts.timestamp,
		},
		Payload: payload,
	}
}
