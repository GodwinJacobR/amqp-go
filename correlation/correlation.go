package correlation

import (
	"context"

	"github.com/google/uuid"
)

const correlationIDKey string = "correlationId"

func GetCorrelationID(ctx context.Context) string {
	newCorrelationID := uuid.NewString()
	if ctx == nil {
		return newCorrelationID
	}
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}

	return newCorrelationID
}
