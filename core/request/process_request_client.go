// +build client

package request

import (
	"context"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// ProcessRequest is called by every request thola receives
func ProcessRequest(ctx context.Context, request Request) (Response, error) {
	logger := log.With().Str("request_id", xid.New().String()).Logger()
	ctx = logger.WithContext(ctx)
	return request.process(ctx)
}
