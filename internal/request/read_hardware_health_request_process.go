//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadHardwareHealthRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	result, err := com.GetHardwareHealthComponent(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hardware health component")
	}

	return &ReadHardwareHealthResponse{
		HardwareHealthComponent: result,
	}, nil
}
