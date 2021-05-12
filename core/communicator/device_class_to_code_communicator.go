package communicator

import (
	"errors"
	"fmt"
	"github.com/inexio/thola/core/tholaerr"
)

func getCodeCommunicator(devClass *deviceClass) (availableCommunicatorFunctions, error) {
	if devClass == nil {
		return nil, errors.New("device class is nil")
	}
	var parent NetworkDeviceCommunicator
	if devClass.parentDeviceClass != nil {
		parent = &(deviceClassCommunicator{devClass.parentDeviceClass})
	}
	classIdentifier := devClass.getName()
	switch classIdentifier {
	case "generic":
		return &genericCommunicator{baseCommunicator{parent}}, nil
	case "ceraos/ip10":
		return &ceraosIP10Communicator{baseCommunicator{parent}}, nil
	case "powerone/acc":
		return &poweroneACCCommunicator{baseCommunicator{parent}}, nil
	case "powerone/pcc":
		return &poweronePCCCommunicator{baseCommunicator{parent}}, nil
	case "ironware":
		return &ironwareCommunicator{baseCommunicator{parent}}, nil
	case "ios":
		return &iosCommunicator{baseCommunicator{parent}}, nil
	case "ekinops":
		return &ekinopsCommunicator{baseCommunicator{parent}}, nil
	case "adva_fsp3kr7":
		return &advaCommunicator{baseCommunicator{parent}}, nil
	case "timos/sas":
		return &timosSASCommunicator{baseCommunicator{parent}}, nil
	case "timos":
		return &timosCommunicator{baseCommunicator{parent}}, nil
	}
	return nil, tholaerr.NewNotFoundError(fmt.Sprintf("no communicator found for device class identifier '%s'", classIdentifier))
}
