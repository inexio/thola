package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/pkg/errors"
	"regexp"
)

type ceraosIP10Communicator struct {
	baseCommunicator
}

func (c *ceraosIP10Communicator) GetIfTable(ctx context.Context) ([]device.Interface, error) {
	subInterfaces, err := c.sub.GetIfTable(ctx)
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
			subInterfaces[i].IfHCInOctets = targetInterface.IfHCInOctets
			subInterfaces[i].IfHCOutOctets = targetInterface.IfHCOutOctets
			subInterfaces[i].IfOperStatus = targetInterface.IfOperStatus
			subInterfaces[i].IfInOctets = targetInterface.IfInOctets
			subInterfaces[i].IfOutOctets = targetInterface.IfOutOctets
			subInterfaces[i].IfInErrors = targetInterface.IfInErrors
			subInterfaces[i].IfOutErrors = targetInterface.IfOutErrors
			break
		}
	}

	return subInterfaces, nil
}
