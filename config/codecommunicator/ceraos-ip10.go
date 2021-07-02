package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/pkg/errors"
	"regexp"
)

type ceraosIP10Communicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of ceraos/ip10 devices.
// These devices need special behavior radio and ethernet interfaces.
func (c *ceraosIP10Communicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	subInterfaces, err := c.parent.GetInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "an unexpected error occurred while trying to get ifTable of sub communicator")
	}

	var targetInterface device.Interface

	regex, err := regexp.Compile("^Ethernet #8$")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build regex")
	}

	for i, inter := range subInterfaces {
		if regex.MatchString(*inter.IfDescr) {
			targetInterface = inter
			copy(subInterfaces[i:], subInterfaces[i+1:])
			subInterfaces = subInterfaces[:len(subInterfaces)-1]
			break
		}
	}

	regex, err = regexp.Compile("^Radio Interface #[0-9]+$")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build regex")
	}

	for i := range subInterfaces {
		if regex.MatchString(*subInterfaces[i].IfDescr) {
			subInterfaces[i].IfOperStatus = targetInterface.IfOperStatus
			subInterfaces[i].IfInOctets = targetInterface.IfInOctets
			subInterfaces[i].IfOutOctets = targetInterface.IfOutOctets
			subInterfaces[i].IfInErrors = targetInterface.IfInErrors
			subInterfaces[i].IfOutErrors = targetInterface.IfOutErrors
			subInterfaces[i].IfHCInOctets = targetInterface.IfHCInOctets
			subInterfaces[i].IfHCOutOctets = targetInterface.IfHCOutOctets

			if subInterfaces[i].IfSpeed != nil {
				speed := *subInterfaces[i].IfSpeed * 1000
				subInterfaces[i].IfSpeed = &speed
			}

			break
		}
	}

	return subInterfaces, nil
}
