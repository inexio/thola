package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
)

type ceraosIP10Communicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of ceraos/ip10 devices.
// These devices need special behavior radio and ethernet interfaces.
func (c *ceraosIP10Communicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {

	subInterfaces, err := c.parent.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	osVersion, err := c.parent.GetOSVersion(ctx)
	var oldOSVersion bool

	// the radio interface of Ceragon IP 10 devices returns a wrong ifSpeed for older os version (< 7)
	// ifSpeed needs to be multiplied by 1000 if os version is < 7
	if err == nil {
		matches := regexp.MustCompile(`^([0-9]+)[Q]*\.`).FindStringSubmatch(osVersion)
		if len(matches) >= 2 {
			majorVersion, err := strconv.Atoi(matches[1])
			if err == nil && majorVersion < 7 {
				oldOSVersion = true
			}
		}
	}

	type config = struct {
		OSRegex                   string
		SourceInterfaceRegex      string
		DestinationInterfaceRegex string
	}

	var configStruct = [2]config{
		{OSRegex: "^[0-9]+[.]", SourceInterfaceRegex: "Ethernet #8", DestinationInterfaceRegex: "Radio Interface #[0-9]+"},
		{OSRegex: "^[0-9]+Q[.]", SourceInterfaceRegex: "Ethernet #5", DestinationInterfaceRegex: "Radio Interface #[0-9]+"},
	}

	for _, c := range configStruct {
		var targetInterface device.Interface

		matches := regexp.MustCompile(c.OSRegex).FindStringSubmatch(osVersion)

		// regex not match -> continue
		if matches == nil {
			continue
		}

		sourceInterfaceRegex, err := regexp.Compile(c.SourceInterfaceRegex)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build regex")
		}

		destinationInterfaceRegex, err := regexp.Compile(c.DestinationInterfaceRegex)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build regex")
		}

		// search the source interface and notice the array integer value
		// remove the source interface from subInterfaces list
		for j, inter := range subInterfaces {
			if inter.IfDescr != nil && sourceInterfaceRegex.MatchString(*inter.IfDescr) {
				targetInterface = inter
				copy(subInterfaces[j:], subInterfaces[j+1:])
				subInterfaces = subInterfaces[:len(subInterfaces)-1]
				break
			}
		}

		// continue if no targetInterface found
		if targetInterface.IfDescr == nil {
			continue
		}

		// copy source interface properties to destination interface
		for j := range subInterfaces {
			if subInterfaces[j].IfDescr != nil && destinationInterfaceRegex.MatchString(*subInterfaces[j].IfDescr) {
				subInterfaces[j].IfOperStatus = targetInterface.IfOperStatus
				subInterfaces[j].IfInOctets = targetInterface.IfInOctets
				subInterfaces[j].IfOutOctets = targetInterface.IfOutOctets
				subInterfaces[j].IfInErrors = targetInterface.IfInErrors
				subInterfaces[j].IfOutErrors = targetInterface.IfOutErrors
				subInterfaces[j].IfHCInOctets = targetInterface.IfHCInOctets
				subInterfaces[j].IfHCOutOctets = targetInterface.IfHCOutOctets

				// ifSpeed needs to be multiplied by 1000 if os version is < 7
				if oldOSVersion && subInterfaces[j].IfSpeed != nil {
					speed := *subInterfaces[j].IfSpeed * 1000
					subInterfaces[j].IfSpeed = &speed
				}
				break
			}
		}
	}

	return filterInterfaces(ctx, subInterfaces, filter)
}
