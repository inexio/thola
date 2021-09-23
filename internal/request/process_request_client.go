// +build client

package request

import (
	"context"
)

type ctxKey byte

const requestIDKey ctxKey = iota + 1

// ProcessRequest is called by every request thola receives
func ProcessRequest(ctx context.Context, request Request) (Response, error) {
	return request.process(ctx)
}

// NewContextWithRequestID returns a new context with the request ID
func NewContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext returns the request ID from the context
func RequestIDFromContext(ctx context.Context) (string, bool) {
	properties, ok := ctx.Value(requestIDKey).(string)
	return properties, ok
}
