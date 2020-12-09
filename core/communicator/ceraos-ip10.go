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

	for i, inter := range subInterfaces {
		ok, err := regexp.MatchString("^Ethernet #8$", *inter.IfDescr)
		if err != nil {
			return nil, errors.Wrap(err, "an unexpected error occurred while trying to match a regexp")
		}

		if ok {
			targetInterface = inter
			copy(subInterfaces[i:], subInterfaces[i+1:])
			subInterfaces = subInterfaces[:len(subInterfaces)-1]
			break
		}
	}

	for i := range subInterfaces {
		ok, err := regexp.MatchString("^Radio Interface #[0-9]+$", *subInterfaces[i].IfDescr)
		if err != nil {
			return nil, errors.Wrap(err, "an unexpected error occurred while trying to match a regexp")
		}

		if ok {
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
