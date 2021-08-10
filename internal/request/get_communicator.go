// +build !client

package request

import (
	"context"
	"github.com/inexio/thola/internal/communicator/communicator"
	"github.com/inexio/thola/internal/communicator/create"
	"github.com/inexio/thola/internal/database"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// GetCommunicator returns a NetworkDeviceCommunicator for the given device.
func GetCommunicator(ctx context.Context, baseRequest BaseRequest) (communicator.Communicator, error) {
	db, err := database.GetDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get DB")
	}

	var invalidCache bool
	deviceProperties, err := db.GetDeviceProperties(ctx, baseRequest.DeviceData.IPAddress)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) {
			return nil, errors.Wrap(err, "failed to get device properties from cache")
		}
		log.Ctx(ctx).Debug().Msg("no device properties found in cache")
		invalidCache = true
	} else {
		logger := log.Ctx(ctx).With().Str("device_class", deviceProperties.Class).Logger()
		ctx := logger.WithContext(ctx)

		log.Ctx(ctx).Debug().Msg("found device properties in cache, starting to validate")
		res, err := create.MatchDeviceClass(ctx, deviceProperties.Class)
		if err != nil {
			return nil, errors.Wrap(err, "failed to match device class")
		}
		if invalidCache = !res; invalidCache {
			log.Ctx(ctx).Debug().Msg("cached device class is invalid")
		} else {
			log.Ctx(ctx).Debug().Msg("cached device class is valid")
		}
	}
	if invalidCache {
		identifyRequest := IdentifyRequest{BaseRequest: baseRequest}
		res, err := identifyRequest.process(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to run identify")
		}
		deviceProperties = res.(*IdentifyResponse).Device
	}
	ctx = device.NewContextWithDeviceProperties(ctx, deviceProperties)

	com, err := create.GetNetworkDeviceCommunicator(ctx, deviceProperties.Class)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get communicator for os '%s'", deviceProperties.Class)
	}

	err = com.UpdateConnection(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update connection")
	}

	return com, nil
}
