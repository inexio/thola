// +build !client

package request

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"thola/core/communicator"
	"thola/core/device"
	"thola/core/tholaerr"
)

// GetCommunicator returns a NetworkDeviceCommunicator for the given device.
func GetCommunicator(ctx context.Context, baseRequest BaseRequest) (communicator.NetworkDeviceCommunicator, error) {
	db, err := getDB()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get DB")
		return nil, errors.Wrap(err, "failed to get DB")
	}

	var invalidCache bool
	identifyData, err := db.GetIdentifyData(baseRequest.DeviceData.IPAddress)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get connection data from cache")
			return nil, errors.Wrap(err, "failed to get connection data from cache")
		}
		invalidCache = true
	} else {
		res, err := communicator.MatchDeviceClass(ctx, identifyData.Class)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to match device class")
			return nil, errors.Wrap(err, "failed to match device class")
		}
		invalidCache = !res
	}
	if invalidCache {
		identifyRequest := IdentifyRequest{BaseRequest: baseRequest}
		res, err := identifyRequest.process(ctx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to run identify")
			return nil, errors.Wrap(err, "failed to run identify")
		}
		identifyData = res.(*IdentifyResponse)
	}
	ctx = device.NewContextWithDeviceProperties(ctx, identifyData.Device)

	com, err := communicator.CreateNetworkDeviceCommunicator(ctx, identifyData.Class)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get communicator for os '%s'", identifyData.Device.Class)
	}
	return com, nil
}
