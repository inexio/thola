// +build !client

package request

import (
	"context"
	"github.com/inexio/thola/internal/communicator/create"
	"github.com/inexio/thola/internal/database"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (r *IdentifyRequest) process(ctx context.Context) (Response, error) {
	log.Ctx(ctx).Trace().Msg("starting identify")

	response, err := r.identify(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "identify failed")
	}

	db, err := database.GetDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get DB")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok {
		return nil, errors.New("no connection data found in context")
	}

	err = db.SetDeviceProperties(ctx, r.DeviceData.IPAddress, response.Device)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save device info to cache")
	}

	err = db.SetConnectionData(ctx, r.DeviceData.IPAddress, con.GetIdealConnectionData())
	if err != nil {
		return nil, errors.Wrap(err, "failed to save connection data to cache")
	}

	return response, nil
}

func (r *IdentifyRequest) identify(ctx context.Context) (*IdentifyResponse, error) {
	com, err := create.IdentifyNetworkDeviceCommunicator(ctx)
	if err != nil {
		return nil, err
	}

	var response IdentifyResponse
	response.Class = com.GetIdentifier()

	response.Properties, err = com.GetIdentifyProperties(ctx)
	if err != nil {
		return &response, err
	}
	return &response, nil
}
