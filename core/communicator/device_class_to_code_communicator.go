package communicator

import (
	"fmt"
	"github.com/inexio/thola/core/tholaerr"
)

func getCodeCommunicator(classIdentifier string, rel *relatedNetworkDeviceCommunicators) (availableCommunicatorFunctions, error) {
	switch classIdentifier {
	case "generic":
		return &genericCommunicator{baseCommunicator{rel}}, nil
	case "ceraos/ip10":
		return &ceraosIP10Communicator{baseCommunicator{rel}}, nil
	case "powerone/acc":
		return &poweroneACCCommunicator{baseCommunicator{rel}}, nil
	case "powerone/pcc":
		return &poweronePCCCommunicator{baseCommunicator{rel}}, nil
	}
	return nil, tholaerr.NewNotFoundError(fmt.Sprintf("no communicator found for device class identifier '%s'", classIdentifier))
}
