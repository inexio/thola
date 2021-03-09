package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/tholaerr"
)

type baseCommunicator struct {
	*relatedNetworkDeviceCommunicators
}

type relatedNetworkDeviceCommunicators struct {
	head NetworkDeviceCommunicator
	sub  NetworkDeviceCommunicator
}

func (c *baseCommunicator) GetVendor(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetModel(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetModelSeries(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSerialNumber(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetOSVersion(_ context.Context) (string, error) {
	return "", tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetInterfaces(_ context.Context) ([]device.Interface, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetIfTable(_ context.Context) ([]device.Interface, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetCountInterfaces(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetCPUComponentCPULoad(_ context.Context) ([]float64, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetCPUComponentCPUTemperature(_ context.Context) ([]float64, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetMemoryComponentMemoryUsage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetServerComponentProcs(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetServerComponentUsers(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetDiskComponentStorages(_ context.Context) ([]device.DiskComponentStorage, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentBatteryAmperage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentBatteryCapacity(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentBatteryCurrent(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentBatteryRemainingTime(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentBatteryTemperature(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentBatteryVoltage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentCurrentLoad(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentMainsVoltageApplied(_ context.Context) (bool, error) {
	return false, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentRectifierCurrent(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentAgents(_ context.Context) ([]device.SBCComponentAgent, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentRealms(_ context.Context) ([]device.SBCComponentRealm, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetUPSComponentSystemVoltage(_ context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentGlobalCallPerSecond(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentGlobalConcurrentSessions(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentActiveLocalContacts(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentTranscodingCapacity(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentLicenseCapacity(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentSystemRedundancy(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetHardwareHealthComponentFans(_ context.Context) ([]device.HardwareHealthComponentFan, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetHardwareHealthComponentPowerSupply(_ context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	return nil, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (c *baseCommunicator) GetSBCComponentSystemHealthScore(_ context.Context) (int, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}
