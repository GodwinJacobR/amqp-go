package amqp_go

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	queue     = "queue"
	exchange  = "exchange"
	result    = "result"
	eventName = "event_name"
)

var (
	eventPublishSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_publish_succeeded",
			Help: "Count of AMQP events that were published successfully",
		}, []string{exchange, eventName},
	)

	eventPublishFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_publish_failed",
			Help: "Count of AMQP events that could not be published",
		}, []string{exchange, eventName},
	)
)

func EventPublishSucceeded(exchange string, eventName string) {
	eventPublishSuccessCounter.WithLabelValues(exchange, eventName).Inc()
}

func EventPublishFailed(exchange string, eventName string) {
	eventPublishFailedCounter.WithLabelValues(exchange, eventName).Inc()
}

func InitMetrics(registerer prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		eventPublishSuccessCounter,
		eventPublishFailedCounter,
	}
	for _, collector := range collectors {
		err := registerer.Register(collector)
		if err != nil {
			return err
		}
	}
	return nil
}
