// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"thola/core/communicator"
	"thola/core/network"
)

func (r *IdentifyRequest) process(ctx context.Context) (Response, error) {
	log.Ctx(ctx).Trace().Msg("starting identify")

	response, err := r.identify(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "identify failed")
	}

	db, err := getDB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get DB")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok {
		return nil, errors.New("no connection data found in context")
	}

	err = db.SetIdentifyData(r.DeviceData.IPAddress, con.GetIdealConnectionData(), response)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save connection data to cache")
	}

	return response, nil
}

func (r *IdentifyRequest) identify(ctx context.Context) (*IdentifyResponse, error) {
	com, err := communicator.IdentifyNetworkDeviceCommunicator(ctx)
	if err != nil {
		return nil, err
	}

	var response IdentifyResponse
	response.Class = com.GetDeviceClass()

	response.Properties, err = com.GetIdentifyProperties(ctx)
	if err != nil {
		return &response, err
	}
	return &response, nil
}
