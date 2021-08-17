package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/communicator/communicator"
	"github.com/inexio/thola/internal/communicator/filter"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
)

type codeCommunicator struct {
	deviceClass communicator.Communicator
	parent      communicator.Communicator
}

// GetCodeCommunicator returns the code communicator for the given device class
func GetCodeCommunicator(deviceClass communicator.Communicator, parentNetworkDeviceCommunicator communicator.Communicator) (communicator.Functions, error) {
	if deviceClass == nil {
		return nil, errors.New("device class is empty")
	}
	var base = codeCommunicator{
		deviceClass: deviceClass,
		parent:      parentNetworkDeviceCommunicator,
	}
	classIdentifier := deviceClass.GetIdentifier()
	switch classIdentifier {
	case "ceraos/ip10":
		return &ceraosIP10Communicator{base}, nil
	case "powerone/acc":
		return &poweroneACCCommunicator{base}, nil
	case "powerone/pcc":
		return &poweronePCCCommunicator{base}, nil
	case "ironware":
		return &ironwareCommunicator{base}, nil
	case "ios":
		return &iosCommunicator{base}, nil
	case "ekinops":
		return &ekinopsCommunicator{base}, nil
	case "adva_fsp3kr7":
		return &advaCommunicator{base}, nil
	case "timos/sas":
		return &timosSASCommunicator{base}, nil
	case "timos":
		return &timosCommunicator{base}, nil
	case "junos":
		return &junosCommunicator{base}, nil
	}
	return nil, tholaerr.NewNotFoundError(fmt.Sprintf("no code communicator found for device class identifier '%s'", classIdentifier))
}

func (c *codeCommunicator) GetVendor(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetModel(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetModelSeries(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSerialNumber(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetOSVersion(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetInterfaces(_ context.Context, _ ...filter.PropertyFilter) ([]device.Interface, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetCountInterfaces(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetCPUComponentCPULoad(_ context.Context) ([]float64, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetCPUComponentCPUTemperature(_ context.Context) ([]float64, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetMemoryComponentMemoryUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetServerComponentProcs(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetServerComponentUsers(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetDiskComponentStorages(_ context.Context) ([]device.DiskComponentStorage, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryAmperage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryCapacity(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryCurrent(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryRemainingTime(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryTemperature(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentBatteryVoltage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentCurrentLoad(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentMainsVoltageApplied(_ context.Context) (bool, error) {
	return false, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentRectifierCurrent(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentAgents(_ context.Context) ([]device.SBCComponentAgent, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentRealms(_ context.Context) ([]device.SBCComponentRealm, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetUPSComponentSystemVoltage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentGlobalCallPerSecond(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentGlobalConcurrentSessions(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentActiveLocalContacts(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentTranscodingCapacity(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentLicenseCapacity(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentSystemRedundancy(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentFans(_ context.Context) ([]device.HardwareHealthComponentFan, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetHardwareHealthComponentPowerSupply(_ context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *codeCommunicator) GetSBCComponentSystemHealthScore(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}
