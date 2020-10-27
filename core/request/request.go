package request

import (
	"context"
	"thola/core/network"
)

// Request is the interface which all requests must implement.
type Request interface {
	validate() error
	getTimeout() *int
	setupConnection(ctx context.Context) (*network.RequestDeviceConnection, error)
	process(ctx context.Context) (Response, error)
	handlePreProcessError(error) (Response, error)
}

// Response is a generic interface that is returned by any Request.
type Response interface {
	GetExitCode() int
}
