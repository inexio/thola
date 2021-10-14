package request

import (
	"context"
	"github.com/inexio/thola/internal/network"
)

// Request is the interface which all requests must implement.
type Request interface {
	// HandlePreProcessError implements request specific error handling (e.g. sets state to UNKNOWN and exit code to 3 in case
	// that the request is a check request).
	// Always call HandlePreProcessError if you want to correctly exit with an error before you call the process function
	// on a request.
	HandlePreProcessError(error) (Response, error)

	validate(ctx context.Context) error
	getTimeout() *int
	setupConnection(ctx context.Context) (*network.RequestDeviceConnection, error)
	process(ctx context.Context) (Response, error)
}

// Response is a generic interface that is returned by any Request.
type Response interface {
	GetExitCode() int
}
