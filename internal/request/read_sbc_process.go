//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadSBCRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	result, err := com.GetSBCComponent(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't get sbc components")
	}

	return &ReadSBCResponse{
		SBC: result,
	}, nil
}
