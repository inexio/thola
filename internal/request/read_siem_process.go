//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
)

func (r *ReadSIEMRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	siem, err := com.GetSIEMComponent(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get siem component")
	}

	return &ReadSIEMResponse{
		SIEM: siem,
	}, nil
}
