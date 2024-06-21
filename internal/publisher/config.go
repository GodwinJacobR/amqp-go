package amqppublisher

import (
	"go.opentelemetry.io/otel/trace"
)

type PublisherConfig struct {
	tracer trace.Tracer
}

type Option func(*PublisherConfig)

func DefaultConfigs() PublisherConfig {
	return PublisherConfig{}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(c *PublisherConfig) {
		c.tracer = tracer
	}
}

func (p *PublisherConfig) GetTracer() trace.Tracer {
	return p.tracer
}
