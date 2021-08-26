package create

import (
	"context"
	"github.com/inexio/thola/internal/communicator/communicator"
	"github.com/inexio/thola/internal/communicator/deviceclass"
	"github.com/inexio/thola/internal/communicator/hierarchy"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
)

var genericHierarchy struct {
	hierarchy.Hierarchy
	sync.Once
}

func initHierarchy(ctx context.Context) error {
	var err error
	genericHierarchy.Do(func() {
		genericHierarchy.Hierarchy, err = deviceclass.GetHierarchy()
		log.Ctx(ctx).Debug().Msg("device configurations initialized")
	})
	if err != nil {
		return errors.Wrap(err, "failed to build initial hierarchy")
	}
	if genericHierarchy.NetworkDeviceCommunicator == nil {
		return errors.New("hierarchy isn't initialized")
	}
	return nil
}

// GetNetworkDeviceCommunicator returns the network device communicator for the given identifier
func GetNetworkDeviceCommunicator(ctx context.Context, identifier string) (communicator.Communicator, error) {
	err := initHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	var ok bool
	var hier hierarchy.Hierarchy
	var currentIdentifier string
	configIdentifiers := strings.Split(identifier, "/")

	if configIdentifiers[0] == "generic" {
		return genericHierarchy.NetworkDeviceCommunicator, nil
	}

	currentIdentifier = configIdentifiers[0]
	hier, ok = genericHierarchy.Children[currentIdentifier]
	if !ok {
		return nil, errors.New("hierarchy does not exist")
	}

	for i, ident := range configIdentifiers {
		if i == 0 {
			continue
		}
		currentIdentifier += "/" + ident
		hier, ok = hier.Children[currentIdentifier]
		if !ok {
			return nil, errors.New("device class does not exist")
		}
	}

	return hier.NetworkDeviceCommunicator, nil
}

// IdentifyNetworkDeviceCommunicator identifies a devices and creates a network device communicator.
func IdentifyNetworkDeviceCommunicator(ctx context.Context) (communicator.Communicator, error) {
	err := initHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	setIdentifyConnectionSettings(ctx)

	comm, err := identifyDeviceRecursive(ctx, genericHierarchy.Children, true)
	if err != nil {
		if tholaerr.IsNotFoundError(err) {
			return genericHierarchy.NetworkDeviceCommunicator, nil
		}
		return nil, errors.Wrap(err, "error occurred while identifying device class")
	}

	return comm, nil
}

func identifyDeviceRecursive(ctx context.Context, children map[string]hierarchy.Hierarchy, considerPriority bool) (communicator.Communicator, error) {
	var tryToMatchLastDeviceClasses map[string]hierarchy.Hierarchy

	for n, hier := range children {
		if considerPriority && hier.TryToMatchLast {
			if tryToMatchLastDeviceClasses == nil {
				tryToMatchLastDeviceClasses = make(map[string]hierarchy.Hierarchy)
			}
			tryToMatchLastDeviceClasses[n] = hier
			continue
		}

		logger := log.Ctx(ctx).With().Str("device_class", hier.NetworkDeviceCommunicator.GetIdentifier()).Logger()
		ctx = logger.WithContext(ctx)
		log.Ctx(ctx).Debug().Msgf("starting class match (%s)", hier.NetworkDeviceCommunicator.GetIdentifier())
		match, err := hier.NetworkDeviceCommunicator.Match(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error while trying to match device class: "+hier.NetworkDeviceCommunicator.GetIdentifier())
		}

		if match {
			log.Ctx(ctx).Debug().Msg("device class matched")
			if hier.Children != nil {
				subDeviceClass, err := identifyDeviceRecursive(ctx, hier.Children, true)
				if err != nil {
					if tholaerr.IsNotFoundError(err) {
						return hier.NetworkDeviceCommunicator, nil
					}
					return nil, errors.Wrapf(err, "error occurred while trying to identify sub device class for device class '%s'", hier.NetworkDeviceCommunicator.GetIdentifier())
				}
				return subDeviceClass, nil
			}
			return hier.NetworkDeviceCommunicator, nil
		}
		log.Ctx(ctx).Debug().Msg("device class did not match")
	}
	if tryToMatchLastDeviceClasses != nil {
		deviceClass, err := identifyDeviceRecursive(ctx, tryToMatchLastDeviceClasses, false)
		if err != nil {
			if !tholaerr.IsNotFoundError(err) {
				return nil, err
			}
		} else {
			return deviceClass, nil
		}
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok {
		return nil, errors.New("no connection data found in context")
	}

	// return generic device class
	if (con.SNMP == nil || con.SNMP.SnmpClient.HasSuccessfulCachedRequest()) && (con.HTTP == nil || con.HTTP.HTTPClient.HasSuccessfulCachedRequest()) {
		return nil, errors.New("no network requests to device succeeded")
	}
	return nil, tholaerr.NewNotFoundError("no device class matched")
}

// MatchDeviceClass checks if the device class in the context matches the given identifier.
func MatchDeviceClass(ctx context.Context, identifier string) (bool, error) {
	comm, err := GetNetworkDeviceCommunicator(ctx, identifier)
	if err != nil {
		return false, errors.Wrap(err, "error during GetDeviceClasses")
	}

	setIdentifyConnectionSettings(ctx)

	return comm.Match(ctx)
}

func setIdentifyConnectionSettings(ctx context.Context) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if ok && con.SNMP != nil && (con.RawConnectionData.SNMP.MaxRepetitions == nil || *con.RawConnectionData.SNMP.MaxRepetitions == 0) {
		con.SNMP.SnmpClient.SetMaxRepetitions(1)
	}
}
