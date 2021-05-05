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
	case "ironware":
		return &ironwareCommunicator{baseCommunicator{rel}}, nil
	case "ios":
		return &iosCommunicator{baseCommunicator{rel}}, nil
	case "ekinops":
		return &ekinopsCommunicator{baseCommunicator{rel}}, nil
	case "adva_fsp3kr7":
		return &advaCommunicator{baseCommunicator{rel}}, nil
	case "timos/sas":
		return &timosSASCommunicator{baseCommunicator{rel}}, nil
	case "timos":
		return &timosCommunicator{baseCommunicator{rel}}, nil
	}
	return nil, tholaerr.NewNotFoundError(fmt.Sprintf("no communicator found for device class identifier '%s'", classIdentifier))
}

/*
func addCodePropertyReader(class string, reader string)
var codePropertyReaderMappings = make(map[string]map[string]propertyReader)
var copeGroupPropertyReaderMappings = map[string]map[string]groupPropertyReader{
	"ceraos/ip10": {
		"IfTable": &ceraosIP10GetIfTableGPR{},
	},
}

*/
