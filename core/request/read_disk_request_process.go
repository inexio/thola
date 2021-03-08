// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadDiskRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	result, err := com.GetDiskComponent(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't get disk components")
	}

	return &ReadDiskResponse{
		Disk: result,
	}, nil
}
