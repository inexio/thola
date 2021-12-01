//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadHighAvailabilityRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	ha, err := com.GetHighAvailabilityComponent(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get high availability component")
	}

	return &ReadHighAvailabilityResponse{
		HighAvailabilityComponent: ha,
	}, nil
}
