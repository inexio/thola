package communicator

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"thola/core/network"
	"thola/core/tholaerr"
)

// CreateNetworkDeviceCommunicator creates a communicator.
func CreateNetworkDeviceCommunicator(ctx context.Context, deviceClassIdentifier string) (NetworkDeviceCommunicator, error) {
	deviceClass, err := getDeviceClass(deviceClassIdentifier)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get device class")
		return nil, errors.Wrap(err, "error during GetDeviceClasses")
	}
	return createCommunicatorByDeviceClass(ctx, deviceClass)
}

// IdentifyNetworkDeviceCommunicator identifies a devices and creates a network device communicator
func IdentifyNetworkDeviceCommunicator(ctx context.Context) (NetworkDeviceCommunicator, error) {
	deviceClass, err := identifyDeviceClass(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to identify device class")
		return nil, errors.Wrap(err, "error during IdentifyDeviceClass")
	}
	com, err := createCommunicatorByDeviceClass(ctx, deviceClass)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create communicator for device class '%s'", deviceClass.getName())
	}
	return com, nil
}

// MatchDeviceClass checks if the device class in the context matches the given identifier
func MatchDeviceClass(ctx context.Context, identifier string) (bool, error) {
	deviceClass, err := getDeviceClass(identifier)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get device class")
		return false, errors.Wrap(err, "error during GetDeviceClasses")
	}
	return deviceClass.matchDevice(ctx)
}

// createCommunicatorByDeviceClass creates a communicator based on a device class.
func createCommunicatorByDeviceClass(ctx context.Context, deviceClass *deviceClass) (NetworkDeviceCommunicator, error) {
	maxRepetitions, err := deviceClass.getSNMPMaxRepetitions()
	if err != nil {
		return nil, errors.Wrap(err, "device class does not have maxrepetitions")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if ok && con.SNMP != nil {
		con.SNMP.SnmpClient.SetMaxRepetitions(maxRepetitions)
	}

	return createCommunicatorByDeviceClassRecursive(deviceClass, nil)
}

func createCommunicatorByDeviceClassRecursive(deviceClass *deviceClass, headCommunicator NetworkDeviceCommunicator) (NetworkDeviceCommunicator, error) {
	var ndCommunicator networkDeviceCommunicator

	if headCommunicator == nil {
		headCommunicator = &ndCommunicator
	}
	ndCommunicator.relatedNetworkDeviceCommunicators = &relatedNetworkDeviceCommunicators{
		head: headCommunicator,
	}

	ndCommunicator.deviceClassCommunicator = &deviceClassCommunicator{
		baseCommunicator: baseCommunicator{
			relatedNetworkDeviceCommunicators: ndCommunicator.relatedNetworkDeviceCommunicators,
		},
		deviceClass: deviceClass,
	}

	parentDeviceClass, err := deviceClass.getParentDeviceClass()
	if err != nil {
		if !tholaerr.IsNotFoundError(err) {
			return nil, errors.Wrap(err, "an unexpected error occurred while trying to get parent device class")
		}
		ndCommunicator.sub = nil
	} else {
		subCommunicator, err := createCommunicatorByDeviceClassRecursive(parentDeviceClass, headCommunicator)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create communicator for parent device class")
		}
		ndCommunicator.sub = subCommunicator
	}

	codeCommunicator, err := getCodeCommunicator(deviceClass.getName(), ndCommunicator.relatedNetworkDeviceCommunicators)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) {
			return nil, errors.Wrap(err, "an unexpected error occurred while trying to map os to communicator")
		}
		ndCommunicator.codeCommunicator = nil
	} else {
		ndCommunicator.codeCommunicator = codeCommunicator
	}

	return &ndCommunicator, nil
}
