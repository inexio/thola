//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/pkg/errors"
	"strings"
)

func (r *ReadInterfacesRequest) process(ctx context.Context) (Response, error) {
	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	var filter []groupproperty.Filter
	if len(r.Values) > 0 {
		var values [][]string
		for _, fil := range r.Values {
			values = append(values, strings.Split(fil, "/"))
		}
		filter = append(filter, groupproperty.GetExclusiveValueFilter(values))
	}

	result, err := com.GetInterfaces(ctx, filter...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interfaces")
	}

	return &ReadInterfacesResponse{
		Interfaces: result,
	}, nil
}
