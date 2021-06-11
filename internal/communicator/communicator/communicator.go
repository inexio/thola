package communicator

import (
	"context"
	"github.com/inexio/thola/internal/communicator/component"
	"github.com/inexio/thola/internal/device"
)

// Communicator represents a communicator for a device.
type Communicator interface {

	// GetIdentifier returns the identifier of the class of a network device.
	GetIdentifier() string

	// GetAvailableComponents returns the components available for a network device.
	GetAvailableComponents() []string

	// HasComponent checks whether the specified component is available.
	HasComponent(component component.Component) bool

	// Match checks if the device matches the device class
	Match(ctx context.Context) (bool, error)

	// GetIdentifyProperties returns the identify properties of a device like vendor, model...
	GetIdentifyProperties(ctx context.Context) (device.Properties, error)

	// GetCPUComponent returns the cpu component of a device if available.
	GetCPUComponent(ctx context.Context) (device.CPUComponent, error)

	// GetUPSComponent returns the ups component of a device if available.
	GetUPSComponent(ctx context.Context) (device.UPSComponent, error)

	// GetSBCComponent returns the sbc component of a device if available.
	GetSBCComponent(ctx context.Context) (device.SBCComponent, error)

	// GetServerComponent returns the sbc component of a device if available.
	GetServerComponent(ctx context.Context) (device.ServerComponent, error)

	// GetDiskComponent returns the disk component of a device if available.
	GetDiskComponent(ctx context.Context) (device.DiskComponent, error)

	// GetHardwareHealthComponent returns the hardware health component of a device if available.
	GetHardwareHealthComponent(ctx context.Context) (device.HardwareHealthComponent, error)

	Functions
}

// Functions represents all overrideable functions which can be used to communicate with a device.
type Functions interface {

	// GetVendor returns the vendor of a device.
	GetVendor(ctx context.Context) (string, error)

	// GetModel returns the model of a device.
	GetModel(ctx context.Context) (string, error)

	// GetModelSeries returns the model series of a device.
	GetModelSeries(ctx context.Context) (string, error)

	// GetSerialNumber returns the serial number of a device.
	GetSerialNumber(ctx context.Context) (string, error)

	// GetOSVersion returns the os version of a device.
	GetOSVersion(ctx context.Context) (string, error)

	// GetInterfaces returns the interfaces of a device.
	GetInterfaces(ctx context.Context) ([]device.Interface, error)

	// GetCountInterfaces returns the count of interfaces of a device.
	GetCountInterfaces(ctx context.Context) (int, error)

	availableCPUCommunicatorFunctions
	availableMemoryCommunicatorFunctions
	availableUPSCommunicatorFunctions
	availableSBCCommunicatorFunctions
	availableServerCommunicatorFunctions
	availableDiskCommunicatorFunctions
	availableHardwareHealthCommunicatorFunctions
}

type availableCPUCommunicatorFunctions interface {

	// GetCPUComponentCPULoad returns the cpu load of the device.
	GetCPUComponentCPULoad(ctx context.Context) ([]float64, error)

	// GetCPUComponentCPUTemperature returns the cpu temperature of the device.
	GetCPUComponentCPUTemperature(ctx context.Context) ([]float64, error)
}

type availableMemoryCommunicatorFunctions interface {

	// GetMemoryComponentMemoryUsage returns the memory usage of the device.
	GetMemoryComponentMemoryUsage(ctx context.Context) (float64, error)
}

type availableDiskCommunicatorFunctions interface {

	// GetDiskComponentStorages returns the storages of the device.
	GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error)
}

type availableUPSCommunicatorFunctions interface {

	// GetUPSComponentAlarmLowVoltageDisconnect returns the low voltage disconnect alarm of the ups device.
	GetUPSComponentAlarmLowVoltageDisconnect(ctx context.Context) (int, error)

	// GetUPSComponentBatteryAmperage returns the battery amperage of the ups device.
	GetUPSComponentBatteryAmperage(ctx context.Context) (float64, error)

	// GetUPSComponentBatteryCapacity returns the battery capacity of the ups device.
	GetUPSComponentBatteryCapacity(ctx context.Context) (float64, error)

	// GetUPSComponentBatteryCurrent returns the current battery of the ups device.
	GetUPSComponentBatteryCurrent(ctx context.Context) (float64, error)

	// GetUPSComponentBatteryRemainingTime returns the battery remaining time of the ups device.
	GetUPSComponentBatteryRemainingTime(ctx context.Context) (float64, error)

	// GetUPSComponentBatteryTemperature returns the battery temperature of the ups device.
	GetUPSComponentBatteryTemperature(ctx context.Context) (float64, error)

	// GetUPSComponentBatteryVoltage returns the battery voltage of the ups device.
	GetUPSComponentBatteryVoltage(ctx context.Context) (float64, error)

	// GetUPSComponentCurrentLoad returns the current load of the ups device.
	GetUPSComponentCurrentLoad(ctx context.Context) (float64, error)

	// GetUPSComponentMainsVoltageApplied returns if the main voltage of the ups device is applied.
	GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error)

	// GetUPSComponentRectifierCurrent returns the current rectifier of the ups device.
	GetUPSComponentRectifierCurrent(ctx context.Context) (float64, error)

	// GetUPSComponentSystemVoltage returns the system voltage of the ups device.
	GetUPSComponentSystemVoltage(ctx context.Context) (float64, error)
}

type availableServerCommunicatorFunctions interface {

	// GetServerComponentProcs returns the process count of the device.
	GetServerComponentProcs(ctx context.Context) (int, error)

	// GetServerComponentUsers returns the user count of the device.
	GetServerComponentUsers(ctx context.Context) (int, error)
}

type availableSBCCommunicatorFunctions interface {

	// GetSBCComponentAgents returns the agents of the sbc device.
	GetSBCComponentAgents(ctx context.Context) ([]device.SBCComponentAgent, error)

	// GetSBCComponentRealms returns the realms of the sbc device.
	GetSBCComponentRealms(ctx context.Context) ([]device.SBCComponentRealm, error)

	// GetSBCComponentGlobalCallPerSecond returns the global calls per second of the sbc device.
	GetSBCComponentGlobalCallPerSecond(ctx context.Context) (int, error)

	// GetSBCComponentGlobalConcurrentSessions returns the global concurrent sessions of the sbc device.
	GetSBCComponentGlobalConcurrentSessions(ctx context.Context) (int, error)

	// GetSBCComponentActiveLocalContacts returns the active local contacts of the sbc device.
	GetSBCComponentActiveLocalContacts(ctx context.Context) (int, error)

	// GetSBCComponentTranscodingCapacity returns the transcoding capacity of the sbc device.
	GetSBCComponentTranscodingCapacity(ctx context.Context) (int, error)

	// GetSBCComponentLicenseCapacity returns the license capacity of the sbc device.
	GetSBCComponentLicenseCapacity(ctx context.Context) (int, error)

	// GetSBCComponentSystemRedundancy returns the system redundancy of the sbc device.
	GetSBCComponentSystemRedundancy(ctx context.Context) (int, error)

	// GetSBCComponentSystemHealthScore returns the system health score of the sbc device.
	GetSBCComponentSystemHealthScore(ctx context.Context) (int, error)
}

type availableHardwareHealthCommunicatorFunctions interface {

	// GetHardwareHealthComponentFans returns the fans of the device.
	GetHardwareHealthComponentFans(ctx context.Context) ([]device.HardwareHealthComponentFan, error)

	// GetHardwareHealthComponentPowerSupply returns the power supply of the device.
	GetHardwareHealthComponentPowerSupply(ctx context.Context) ([]device.HardwareHealthComponentPowerSupply, error)

	// GetHardwareHealthComponentEnvironmentMonitorState returns the environment monitoring state of the device.
	GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (int, error)
}
