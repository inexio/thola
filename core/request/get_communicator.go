// +build !client

package request

import (
	"context"
	"github.com/inexio/thola/core/communicator"
	"github.com/inexio/thola/core/database"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// GetCommunicator returns a NetworkDeviceCommunicator for the given device.
func GetCommunicator(ctx context.Context, baseRequest BaseRequest) (communicator.NetworkDeviceCommunicator, error) {
	db, err := database.GetDB(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get DB")
		return nil, errors.Wrap(err, "failed to get DB")
	}

	var invalidCache bool
	deviceProperties, err := db.GetDeviceProperties(ctx, baseRequest.DeviceData.IPAddress)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get connection data from cache")
			return nil, errors.Wrap(err, "failed to get connection data from cache")
		}
		invalidCache = true
	} else {
		res, err := communicator.MatchDeviceClass(ctx, deviceProperties.Class)
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
		deviceProperties = res.(*IdentifyResponse).Device
	}
	ctx = device.NewContextWithDeviceProperties(ctx, deviceProperties)

	com, err := communicator.CreateNetworkDeviceCommunicator(ctx, deviceProperties.Class)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get communicator for os '%s'", deviceProperties.Class)
	}
	return com, nil
}
