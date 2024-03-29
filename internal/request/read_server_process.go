//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadServerRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	result, err := com.GetServerComponent(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't get server components")
	}

	return &ReadServerResponse{
		Server: result,
	}, nil
}
