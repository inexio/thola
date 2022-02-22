package communicator

import (
	"context"
	"github.com/inexio/thola/internal/component"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
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

	// UpdateConnection updates the device connection with class specific values
	UpdateConnection(ctx context.Context) error

	// GetIdentifyProperties returns the identify properties of a device like vendor, model...
	GetIdentifyProperties(ctx context.Context) (device.Properties, error)

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

	// GetHighAvailabilityComponent returns the hardware health component of a device if available.
	GetHighAvailabilityComponent(ctx context.Context) (device.HighAvailabilityComponent, error)

	// GetSIEMComponent returns the siem socmponent component of a device if available.
	GetSIEMComponent(ctx context.Context) (device.SIEMComponent, error)

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
	GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error)

	// GetCountInterfaces returns the count of interfaces of a device.
	GetCountInterfaces(ctx context.Context) (int, error)

	availableCPUCommunicatorFunctions
	availableMemoryCommunicatorFunctions
	availableUPSCommunicatorFunctions
	availableSBCCommunicatorFunctions
	availableServerCommunicatorFunctions
	availableDiskCommunicatorFunctions
	availableHardwareHealthCommunicatorFunctions
	availableHighAvailabilityCommunicatorFunctions
	availableSIEMCommunicatorFunctions
}

type availableCPUCommunicatorFunctions interface {

	// GetCPUComponentCPULoad returns the cpu load of the device.
	GetCPUComponentCPULoad(ctx context.Context) ([]device.CPU, error)
}

type availableMemoryCommunicatorFunctions interface {

	// GetMemoryComponentMemoryUsage returns the memory usage of the device.
	GetMemoryComponentMemoryUsage(ctx context.Context) ([]device.MemoryPool, error)
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
	GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (device.HardwareHealthComponentState, error)

	// GetHardwareHealthComponentTemperature returns the temperature sensors of the device.
	GetHardwareHealthComponentTemperature(context.Context) ([]device.HardwareHealthComponentTemperature, error)

	// GetHardwareHealthComponentVoltage returns the voltages of the device.
	GetHardwareHealthComponentVoltage(context.Context) ([]device.HardwareHealthComponentVoltage, error)
}

type availableHighAvailabilityCommunicatorFunctions interface {

	// GetHighAvailabilityComponentState returns the HA state.
	GetHighAvailabilityComponentState(ctx context.Context) (device.HighAvailabilityComponentState, error)

	// GetHighAvailabilityComponentRole returns the role of the device in its HA setup.
	GetHighAvailabilityComponentRole(ctx context.Context) (string, error)

	// GetHighAvailabilityComponentNodes returns number of nodes in a HA setup.
	GetHighAvailabilityComponentNodes(ctx context.Context) (int, error)
}

type availableSIEMCommunicatorFunctions interface {

	// GetSIEMComponentLastRecordedMessagesPerSecondNormalizer returns last recorded messages per seconds of the normalizer
	GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx context.Context) (int, error)

	// GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer returns average recorded messages per seconds of the normalizer
	GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(ctx context.Context) (int, error)

	// GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler returns last recorded messages per seconds of the store handler
	GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(ctx context.Context) (int, error)

	// GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler returns average recorded messages per seconds of the store handler
	GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(ctx context.Context) (int, error)

	// GetSIEMComponentServicesCurrentlyDown returns currently down services number
	GetSIEMComponentServicesCurrentlyDown(ctx context.Context) (int, error)

	// GetSIEMComponentSystemVersion returns the siem system version
	GetSIEMComponentSystemVersion(ctx context.Context) (string, error)

	// GetSIEMComponentSIEM returns siem type
	GetSIEMComponentSIEM(ctx context.Context) (string, error)

	// GetSIEMComponentCpuConsumptionCollection returns siem type
	GetSIEMComponentCpuConsumptionCollection(ctx context.Context) (float64, error)

	// GetSIEMComponentCpuConsumptionNormalization returns siem type
	GetSIEMComponentCpuConsumptionNormalization(ctx context.Context) (float64, error)

	// GetSIEMComponentCpuConsumptionEnrichment returns siem type
	GetSIEMComponentCpuConsumptionEnrichment(ctx context.Context) (float64, error)

	// GetSIEMComponentCpuConsumptionIndexing returns siem type
	GetSIEMComponentCpuConsumptionIndexing(ctx context.Context) (float64, error)

	// GetSIEMComponentCpuConsumptionDashboardAlerts returns siem type
	GetSIEMComponentCpuConsumptionDashboardAlerts(ctx context.Context) (float64, error)
}
