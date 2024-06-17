package amqppublisher

import (
	"go.opentelemetry.io/otel/trace"
)

type publisherConfig struct {
	tracer trace.Tracer
}

type Option func(*publisherConfig)

func WithTracer(tracer trace.Tracer) Option {
	return func(c *publisherConfig) {
		c.tracer = tracer
	}
}
