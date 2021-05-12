package communicator

import (
	"context"
	"github.com/pkg/errors"
)

// CreateNetworkDeviceCommunicator creates a communicator.
func CreateNetworkDeviceCommunicator(ctx context.Context, deviceClassIdentifier string) (NetworkDeviceCommunicator, error) {
	devClass, err := getDeviceClass(deviceClassIdentifier)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetDeviceClasses")
	}

	return devClass.getNetworkDeviceCommunicator(ctx)
}

// IdentifyNetworkDeviceCommunicator identifies a devices and creates a network device communicator.
func IdentifyNetworkDeviceCommunicator(ctx context.Context) (NetworkDeviceCommunicator, error) {
	devClass, err := identifyDeviceClass(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error during IdentifyDeviceClass")
	}

	return devClass.getNetworkDeviceCommunicator(ctx)
}

// MatchDeviceClass checks if the device class in the context matches the given identifier.
func MatchDeviceClass(ctx context.Context, identifier string) (bool, error) {
	deviceClass, err := getDeviceClass(identifier)
	if err != nil {
		return false, errors.Wrap(err, "error during GetDeviceClasses")
	}
	return deviceClass.matchDevice(ctx)
}
