// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadCPULoadRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	result, err := com.GetCPUComponentCPULoad(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't get cpu load")
	}

	return &ReadCPULoadResponse{
		CPULoad: result,
	}, nil
}
