package communicator

import (
	"context"
	"github.com/inexio/thola/internal/component"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// CreateNetworkDeviceCommunicator creates a network device communicator which combines a device class communicator and code communicator
func CreateNetworkDeviceCommunicator(deviceClassCommunicator Communicator, codeCommunicator Functions) Communicator {
	return &networkDeviceCommunicator{
		deviceClassCommunicator: deviceClassCommunicator,
		codeCommunicator:        codeCommunicator,
	}
}

type networkDeviceCommunicator struct {
	deviceClassCommunicator Communicator
	codeCommunicator        Functions
}

func (c *networkDeviceCommunicator) GetIdentifier() string {
	return c.deviceClassCommunicator.GetIdentifier()
}

// GetAvailableComponents returns the available Components for the device.
func (c *networkDeviceCommunicator) GetAvailableComponents() []string {
	return c.deviceClassCommunicator.GetAvailableComponents()
}

// HasComponent checks whether the specified component is available.
func (c *networkDeviceCommunicator) HasComponent(component component.Component) bool {
	return c.deviceClassCommunicator.HasComponent(component)
}

func (c *networkDeviceCommunicator) Match(ctx context.Context) (bool, error) {
	return c.deviceClassCommunicator.Match(ctx)
}

func (c *networkDeviceCommunicator) UpdateConnection(ctx context.Context) error {
	return c.deviceClassCommunicator.UpdateConnection(ctx)
}

func (c *networkDeviceCommunicator) GetIdentifyProperties(ctx context.Context) (device.Properties, error) {
	dev := device.Device{
		Class:      c.GetIdentifier(),
		Properties: device.Properties{},
	}

	vendor, err := c.GetVendor(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get vendor")
		}
	} else {
		dev.Properties.Vendor = &vendor
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	model, err := c.GetModel(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model")
		}
	} else {
		dev.Properties.Model = &model
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	modelSeries, err := c.GetModelSeries(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model series")
		}
	} else {
		dev.Properties.ModelSeries = &modelSeries
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	serialNumber, err := c.GetSerialNumber(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get serial number")
		}
	} else {
		dev.Properties.SerialNumber = &serialNumber
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	osVersion, err := c.GetOSVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get os version")
		}
	} else {
		dev.Properties.OSVersion = &osVersion
	}

	return dev.Properties, nil
}

func (c *networkDeviceCommunicator) GetDiskComponent(ctx context.Context) (device.DiskComponent, error) {
	if !c.HasComponent(component.Disk) {
		return device.DiskComponent{}, tholaerr.NewComponentNotFoundError("no disk component available for this device")
	}

	var disk device.DiskComponent

	empty := true

	storages, err := c.GetDiskComponentStorages(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.DiskComponent{}, errors.Wrap(err, "error occurred during get disk component storages")
		}
	} else {
		disk.Storages = storages
		empty = false
	}

	if empty {
		return device.DiskComponent{}, tholaerr.NewNotFoundError("no disk data available")
	}

	return disk, nil
}

func (c *networkDeviceCommunicator) GetUPSComponent(ctx context.Context) (device.UPSComponent, error) {
	if !c.HasComponent(component.UPS) {
		return device.UPSComponent{}, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	var ups device.UPSComponent
	empty := true

	alarmLowVoltage, err := c.GetUPSComponentAlarmLowVoltageDisconnect(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get alarm")
		}
	} else {
		ups.AlarmLowVoltageDisconnect = &alarmLowVoltage
		empty = false
	}

	batteryAmperage, err := c.GetUPSComponentBatteryAmperage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery amperage")
		}
	} else {
		ups.BatteryAmperage = &batteryAmperage
		empty = false
	}

	batteryCapacity, err := c.GetUPSComponentBatteryCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryCapacity = &batteryCapacity
		empty = false
	}

	batteryCurrent, err := c.GetUPSComponentBatteryCurrent(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryCurrent = &batteryCurrent
		empty = false
	}

	batteryRemainingTime, err := c.GetUPSComponentBatteryRemainingTime(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryRemainingTime = &batteryRemainingTime
		empty = false
	}

	batteryTemperature, err := c.GetUPSComponentBatteryTemperature(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery temperature")
		}
	} else {
		ups.BatteryTemperature = &batteryTemperature
		empty = false
	}

	batteryVoltage, err := c.GetUPSComponentBatteryVoltage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery voltage")
		}
	} else {
		ups.BatteryVoltage = &batteryVoltage
		empty = false
	}

	currentLoad, err := c.GetUPSComponentCurrentLoad(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get current load")
		}
	} else {
		ups.CurrentLoad = &currentLoad
		empty = false
	}

	mainsVoltageApplied, err := c.GetUPSComponentMainsVoltageApplied(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.MainsVoltageApplied = &mainsVoltageApplied
		empty = false
	}

	rectifierCurrent, err := c.GetUPSComponentRectifierCurrent(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.RectifierCurrent = &rectifierCurrent
		empty = false
	}

	systemVoltage, err := c.GetUPSComponentSystemVoltage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.SystemVoltage = &systemVoltage
		empty = false
	}

	if empty {
		return device.UPSComponent{}, tholaerr.NewNotFoundError("no ups data available")
	}
	return ups, nil
}

func (c *networkDeviceCommunicator) GetServerComponent(ctx context.Context) (device.ServerComponent, error) {
	if !c.HasComponent(component.Server) {
		return device.ServerComponent{}, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}

	var server device.ServerComponent

	empty := true

	procs, err := c.GetServerComponentProcs(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.ServerComponent{}, errors.Wrap(err, "error occurred during get server component procs")
		}
	} else {
		server.Procs = &procs
		empty = false
	}

	users, err := c.GetServerComponentUsers(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.ServerComponent{}, errors.Wrap(err, "error occurred during get server component users")
		}
	} else {
		server.Users = &users
		empty = false
	}

	if empty {
		return device.ServerComponent{}, tholaerr.NewNotFoundError("no server data available")
	}

	return server, nil
}

func (c *networkDeviceCommunicator) GetSBCComponent(ctx context.Context) (device.SBCComponent, error) {
	if !c.HasComponent(component.SBC) {
		return device.SBCComponent{}, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	var sbc device.SBCComponent

	empty := true

	agents, err := c.GetSBCComponentAgents(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component agents")
		}
	} else {
		sbc.Agents = agents
		empty = false
	}

	realms, err := c.GetSBCComponentRealms(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component realms")
		}
	} else {
		sbc.Realms = realms
		empty = false
	}

	globalCPS, err := c.GetSBCComponentGlobalCallPerSecond(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component sbc global call per second")
		}
	} else {
		sbc.GlobalCallPerSecond = &globalCPS
		empty = false
	}

	globalConcurrentSessions, err := c.GetSBCComponentGlobalConcurrentSessions(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc global concurrent sessions")
		}
	} else {
		sbc.GlobalConcurrentSessions = &globalConcurrentSessions
		empty = false
	}

	activeLocalContacts, err := c.GetSBCComponentActiveLocalContacts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get active local contacts")
		}
	} else {
		sbc.ActiveLocalContacts = &activeLocalContacts
		empty = false
	}

	transcodingCapacity, err := c.GetSBCComponentTranscodingCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get transcoding capacity")
		}
	} else {
		sbc.TranscodingCapacity = &transcodingCapacity
		empty = false
	}

	licenseCapacity, err := c.GetSBCComponentLicenseCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get license capacity")
		}
	} else {
		sbc.LicenseCapacity = &licenseCapacity
		empty = false
	}

	systemRedundancy, err := c.GetSBCComponentSystemRedundancy(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get system redundancy")
		}
	} else {
		sbc.SystemRedundancy = &systemRedundancy
		empty = false
	}

	systemHealthScore, err := c.GetSBCComponentSystemHealthScore(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get system health score")
		}
	} else {
		sbc.SystemHealthScore = &systemHealthScore
		empty = false
	}

	if empty {
		return device.SBCComponent{}, tholaerr.NewNotFoundError("no sbc data available")
	}

	return sbc, nil
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponent(ctx context.Context) (device.HardwareHealthComponent, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return device.HardwareHealthComponent{}, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	var hardwareHealth device.HardwareHealthComponent

	empty := true

	state, err := c.GetHardwareHealthComponentEnvironmentMonitorState(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get environment monitor states")
		}
	} else {
		hardwareHealth.EnvironmentMonitorState = &state
		empty = false
	}

	fans, err := c.GetHardwareHealthComponentFans(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get fans")
		}
	} else {
		hardwareHealth.Fans = fans
		empty = false
	}

	powerSupply, err := c.GetHardwareHealthComponentPowerSupply(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get power supply")
		}
	} else {
		hardwareHealth.PowerSupply = powerSupply
		empty = false
	}

	temp, err := c.GetHardwareHealthComponentTemperature(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get temperature")
		}
	} else {
		hardwareHealth.Temperature = temp
		empty = false
	}

	volt, err := c.GetHardwareHealthComponentVoltage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get voltage")
		}
	} else {
		hardwareHealth.Voltage = volt
		empty = false
	}

	if empty {
		return device.HardwareHealthComponent{}, tholaerr.NewNotFoundError("no hardware health data available")
	}

	return hardwareHealth, nil
}

func (c *networkDeviceCommunicator) GetHighAvailabilityComponent(ctx context.Context) (device.HighAvailabilityComponent, error) {
	if !c.HasComponent(component.HighAvailability) {
		return device.HighAvailabilityComponent{}, tholaerr.NewComponentNotFoundError("no ha component available for this device")
	}

	var ha device.HighAvailabilityComponent

	empty := true

	state, err := c.GetHighAvailabilityComponentState(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HighAvailabilityComponent{}, errors.Wrap(err, "error occurred during get high availability state")
		}
	} else {
		ha.State = &state
		empty = false
	}

	// if device is in standalone mode, return as there is no high-availability setup running
	if state == device.HighAvailabilityComponentStateStandalone {
		return ha, nil
	}

	role, err := c.GetHighAvailabilityComponentRole(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HighAvailabilityComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		ha.Role = &role
		empty = false
	}

	nodes, err := c.GetHighAvailabilityComponentNodes(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HighAvailabilityComponent{}, errors.Wrap(err, "error occurred during get high availability nodes")
		}
	} else {
		ha.Nodes = &nodes
		empty = false
	}

	if empty {
		return device.HighAvailabilityComponent{}, tholaerr.NewNotFoundError("no high availability data available")
	}

	return ha, nil
}

func (c *networkDeviceCommunicator) GetSIEMComponent(ctx context.Context) (device.SIEMComponent, error) {
	if !c.HasComponent(component.SIEM) {
		return device.SIEMComponent{}, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	var siem device.SIEMComponent

	empty := true

	lrmpsNormalizer, err := c.GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.LastRecordedMessagesPerSecondNormalizer = &lrmpsNormalizer
		empty = false
	}

	armpsNormalizer, err := c.GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.AverageMessagesPerSecondLast5minNormalizer = &armpsNormalizer
		empty = false
	}

	lrmpsHandler, err := c.GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.LastRecordedMessagesPerSecondStoreHandler = &lrmpsHandler
		empty = false
	}

	armpsHandler, err := c.GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.AverageMessagesPerSecondLast5minStoreHandler = &armpsHandler
		empty = false
	}

	servicesDown, err := c.GetSIEMComponentServicesCurrentlyDown(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ServicesCurrentlyDown = &servicesDown
		empty = false
	}

	systemVersion, err := c.GetSIEMComponentSystemVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.SystemVersion = &systemVersion
		empty = false
	}

	siemType, err := c.GetSIEMComponentSIEM(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.SIEM = &siemType
		empty = false
	}

	cpuDashboardAlert, err := c.GetSIEMComponentCpuConsumptionDashboardAlerts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.CpuConsumptionDashboardAlerts = &cpuDashboardAlert
		empty = false
	}

	cpuNormalization, err := c.GetSIEMComponentCpuConsumptionNormalization(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.CpuConsumptionNormalization = &cpuNormalization
		empty = false
	}

	cpuIndexing, err := c.GetSIEMComponentCpuConsumptionIndexing(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.CpuConsumptionIndexing = &cpuIndexing
		empty = false
	}

	cpuCollection, err := c.GetSIEMComponentCpuConsumptionCollection(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.CpuConsumptionCollection = &cpuCollection
		empty = false
	}

	cpuEnrichment, err := c.GetSIEMComponentCpuConsumptionEnrichment(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.CpuConsumptionEnrichment = &cpuEnrichment
		empty = false
	}

	memoryDashboardAlert, err := c.GetSIEMComponentMemoryConsumptionDashboardAlerts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.MemoryConsumptionDashboardAlerts = &memoryDashboardAlert
		empty = false
	}

	memoryNormalization, err := c.GetSIEMComponentMemoryConsumptionNormalization(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.MemoryConsumptionNormalization = &memoryNormalization
		empty = false
	}

	memoryIndexing, err := c.GetSIEMComponentMemoryConsumptionIndexing(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.MemoryConsumptionIndexing = &memoryIndexing
		empty = false
	}

	memoryCollection, err := c.GetSIEMComponentMemoryConsumptionCollection(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.MemoryConsumptionCollection = &memoryCollection
		empty = false
	}

	memoryEnrichment, err := c.GetSIEMComponentMemoryConsumptionEnrichment(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.MemoryConsumptionEnrichment = &memoryEnrichment
		empty = false
	}

	queueDashboardAlert, err := c.GetSIEMComponentQueueDashboardAlerts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.QueueDashboardAlerts = &queueDashboardAlert
		empty = false
	}

	queueNormalization, err := c.GetSIEMComponentQueueNormalization(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.QueueNormalization = &queueNormalization
		empty = false
	}

	queueIndexing, err := c.GetSIEMComponentQueueIndexing(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.QueueIndexing = &queueIndexing
		empty = false
	}

	queueCollection, err := c.GetSIEMComponentQueueCollection(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.QueueCollection = &queueCollection
		empty = false
	}

	queueEnrichment, err := c.GetSIEMComponentQueueEnrichment(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.QueueEnrichment = &queueEnrichment
		empty = false
	}

	activeSearches, err := c.GetSIEMComponentActiveSearchProcesses(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ActiveSearchProcesses = &activeSearches
		empty = false
	}

	diskUsageDA, err := c.GetSIEMComponentDiskUsageDashboardAlerts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.DiskUsageDashboardAlerts = &diskUsageDA
		empty = false
	}

	zfs, err := c.GetSIEMComponentZFSPools(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ZFSPools = zfs
		empty = false
	}

	repos, err := c.GetSIEMComponentRepositories(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.Repositories = repos
		empty = false
	}

	fsVersion, err := c.GetSIEMComponentFabricServerVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerVersion = &fsVersion
		empty = false
	}

	fsIOWait, err := c.GetSIEMComponentFabricServerIOWait(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerIOWait = &fsIOWait
		empty = false
	}

	fsVMswap, err := c.GetSIEMComponentFabricServerVMSwapiness(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerVMSwapiness = &fsVMswap
		empty = false
	}

	fsClusterSize, err := c.GetSIEMComponentFabricServerClusterSize(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerClusterSize = &fsClusterSize
		empty = false
	}

	fsProxyCpu, err := c.GetSIEMComponentFabricServerProxyCpuUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerProxyCpuUsage = &fsProxyCpu
		empty = false
	}

	fsProxyMem, err := c.GetSIEMComponentFabricServerProxyMemoryUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerProxyMemoryUsage = &fsProxyMem
		empty = false
	}

	fsProxyAlCon, err := c.GetSIEMComponentFabricServerProxyNumberOfAliveConnections(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerProxyNumberOfAliveConnections = &fsProxyAlCon
		empty = false
	}

	fsProxyState, err := c.GetSIEMComponentFabricServerProxyState(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerProxyState = &fsProxyState
		empty = false
	}

	fsProxyNodesCount, err := c.GetSIEMComponentFabricServerProxyNodesCount(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerProxyNodesCount = &fsProxyNodesCount
		empty = false
	}

	fsStorageCPU, err := c.GetSIEMComponentFabricServerStorageCPUUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageCpuUsage = &fsStorageCPU
		empty = false
	}

	fsStorageMem, err := c.GetSIEMComponentFabricServerStorageMemoryUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageMemoryUsage = &fsStorageMem
		empty = false
	}

	fsConfiguredCap, err := c.GetSIEMComponentConfiguredCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageConfiguredCapacity = &fsConfiguredCap
		empty = false
	}

	fsAvailCap, err := c.GetSIEMComponentFabricServerStorageAvailableCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageAvailableCapacity = &fsAvailCap
		empty = false
	}

	fsStorageDFSUsed, err := c.GetSIEMComponentFabricServerStorageDFSUsed(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageDfsUsed = &fsStorageDFSUsed
		empty = false
	}

	fsStorageURB, err := c.GetSIEMComponentFabricServerStorageUnderReplicatedBlocks(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageUnderReplicatedBlocks = &fsStorageURB
		empty = false
	}

	fsStorageLiveDataNodes, err := c.GetSIEMComponentFabricServerStorageLiveDataNodes(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerStorageLiveDataNodes = &fsStorageLiveDataNodes
		empty = false
	}

	fsAuthCPU, err := c.GetSIEMComponentFabricServerAuthenticatorCPUUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerAuthenticatorCpuUsage = &fsAuthCPU
		empty = false
	}

	fsAuthMem, err := c.GetSIEMComponentFabricServerAuthenticatorMemoryUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerAuthenticatorMemoryUsage = &fsAuthMem
		empty = false
	}

	fsAuthServiceStatus, err := c.GetSIEMComponentFabricServerAuthenticatorServiceStatus(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerAuthenticatorServiceStatus = &fsAuthServiceStatus
		empty = false
	}

	fsAuthAdminServiceStat, err := c.GetSIEMComponentFabricServerAuthenticatorAdminServiceStatus(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerAuthenticatorAdminServiceStatus = &fsAuthAdminServiceStat
		empty = false
	}

	fsZFSPools, err := c.GetSIEMComponentFabricServerZFSPools(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.FabricServerZFSPools = fsZFSPools
		empty = false
	}

	apiVersion, err := c.GetSIEMComponentAPIServerVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ApiServerVersion = &apiVersion
		empty = false
	}

	apiIOWait, err := c.GetSIEMComponentAPIServerIOWait(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ApiServerIOWait = &apiIOWait
		empty = false
	}

	apiVMSwap, err := c.GetSIEMComponentAPIServerVMSwapiness(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ApiServerVMSwapiness = &apiVMSwap
		empty = false
	}

	apiCPU, err := c.GetSIEMComponentAPIServerCPUUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ApiServerCpuUsage = &apiCPU
		empty = false
	}

	apiMem, err := c.GetSIEMComponentAPIServerMemoryUsage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.ApiServerMemoryUsage = &apiMem
		empty = false
	}

	if empty {
		return device.SIEMComponent{}, tholaerr.NewNotFoundError("no high availability data available")
	}

	return siem, nil
}

func (c *networkDeviceCommunicator) GetVendor(ctx context.Context) (string, error) {
	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetVendor(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetVendor(ctx)
}

func (c *networkDeviceCommunicator) GetModel(ctx context.Context) (string, error) {
	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetModel(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetModel(ctx)
}

func (c *networkDeviceCommunicator) GetModelSeries(ctx context.Context) (string, error) {
	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetModelSeries(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetModelSeries(ctx)
}

func (c *networkDeviceCommunicator) GetSerialNumber(ctx context.Context) (string, error) {
	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSerialNumber(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSerialNumber(ctx)
}

func (c *networkDeviceCommunicator) GetOSVersion(ctx context.Context) (string, error) {
	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetOSVersion(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetOSVersion(ctx)
}

func (c *networkDeviceCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	if !c.HasComponent(component.Interfaces) {
		return nil, tholaerr.NewComponentNotFoundError("no interface component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetInterfaces(ctx, filter...)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetInterfaces(ctx, filter...)
}

func (c *networkDeviceCommunicator) GetCountInterfaces(ctx context.Context) (int, error) {
	if !c.HasComponent(component.Interfaces) {
		return 0, tholaerr.NewComponentNotFoundError("no interface component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetCountInterfaces(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	amount, err := c.deviceClassCommunicator.GetCountInterfaces(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Msg("failed to get count interfaces, trying to get interfaces")
		var interfaces []device.Interface
		interfaces, err = c.GetInterfaces(ctx, groupproperty.GetExclusiveValueFilter([][]string{{"ifIndex"}, {"ifDescr"}}))
		if err != nil {
			return 0, errors.Wrap(err, "count interfaces failed")
		}
		amount = len(interfaces)
	}

	return amount, err
}

func (c *networkDeviceCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]device.CPU, error) {
	if !c.HasComponent(component.CPU) {
		return nil, tholaerr.NewComponentNotFoundError("no cpu component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetCPUComponentCPULoad(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetCPUComponentCPULoad(ctx)
}

func (c *networkDeviceCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) ([]device.MemoryPool, error) {
	if !c.HasComponent(component.Memory) {
		return nil, tholaerr.NewComponentNotFoundError("no memory component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetMemoryComponentMemoryUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetMemoryComponentMemoryUsage(ctx)
}

func (c *networkDeviceCommunicator) GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error) {
	if !c.HasComponent(component.Disk) {
		return nil, tholaerr.NewComponentNotFoundError("no disk component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetDiskComponentStorages(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetDiskComponentStorages(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(ctx context.Context) (int, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentAlarmLowVoltageDisconnect(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentAlarmLowVoltageDisconnect(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryAmperage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentBatteryAmperage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentBatteryAmperage(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryCapacity(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentBatteryCapacity(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentBatteryCapacity(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryCurrent(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentBatteryCurrent(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentBatteryCurrent(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryRemainingTime(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentBatteryRemainingTime(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentBatteryRemainingTime(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryTemperature(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentBatteryTemperature(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentBatteryTemperature(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryVoltage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentBatteryVoltage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentBatteryVoltage(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentCurrentLoad(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentCurrentLoad(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentCurrentLoad(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error) {
	if !c.HasComponent(component.UPS) {
		return false, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentMainsVoltageApplied(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return false, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentMainsVoltageApplied(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentRectifierCurrent(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentRectifierCurrent(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentRectifierCurrent(ctx)
}

func (c *networkDeviceCommunicator) GetUPSComponentSystemVoltage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.UPS) {
		return 0, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetUPSComponentSystemVoltage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetUPSComponentSystemVoltage(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentAgents(ctx context.Context) ([]device.SBCComponentAgent, error) {
	if !c.HasComponent(component.SBC) {
		return nil, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentAgents(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentAgents(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentRealms(ctx context.Context) ([]device.SBCComponentRealm, error) {
	if !c.HasComponent(component.SBC) {
		return nil, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentRealms(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentRealms(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentGlobalCallPerSecond(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentGlobalCallPerSecond(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentGlobalCallPerSecond(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentGlobalConcurrentSessions(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentGlobalConcurrentSessions(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentGlobalConcurrentSessions(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentActiveLocalContacts(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentActiveLocalContacts(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentActiveLocalContacts(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentTranscodingCapacity(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentTranscodingCapacity(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentTranscodingCapacity(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentLicenseCapacity(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentLicenseCapacity(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentLicenseCapacity(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentSystemRedundancy(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentSystemRedundancy(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentSystemRedundancy(ctx)
}

func (c *networkDeviceCommunicator) GetSBCComponentSystemHealthScore(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SBC) {
		return 0, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSBCComponentSystemHealthScore(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSBCComponentSystemHealthScore(ctx)
}

func (c *networkDeviceCommunicator) GetServerComponentProcs(ctx context.Context) (int, error) {
	if !c.HasComponent(component.Server) {
		return 0, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetServerComponentProcs(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetServerComponentProcs(ctx)
}

func (c *networkDeviceCommunicator) GetServerComponentUsers(ctx context.Context) (int, error) {
	if !c.HasComponent(component.Server) {
		return 0, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetServerComponentUsers(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetServerComponentUsers(ctx)
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (device.HardwareHealthComponentState, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return "", tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHardwareHealthComponentEnvironmentMonitorState(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHardwareHealthComponentEnvironmentMonitorState(ctx)
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentFans(ctx context.Context) ([]device.HardwareHealthComponentFan, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return nil, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHardwareHealthComponentFans(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHardwareHealthComponentFans(ctx)
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentPowerSupply(ctx context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return nil, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHardwareHealthComponentPowerSupply(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHardwareHealthComponentPowerSupply(ctx)
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentTemperature(ctx context.Context) ([]device.HardwareHealthComponentTemperature, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return nil, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHardwareHealthComponentTemperature(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHardwareHealthComponentTemperature(ctx)
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentVoltage(ctx context.Context) ([]device.HardwareHealthComponentVoltage, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return nil, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHardwareHealthComponentVoltage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHardwareHealthComponentVoltage(ctx)
}

func (c *networkDeviceCommunicator) GetHighAvailabilityComponentState(ctx context.Context) (device.HighAvailabilityComponentState, error) {
	if !c.HasComponent(component.HighAvailability) {
		return "", tholaerr.NewComponentNotFoundError("no ha component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHighAvailabilityComponentState(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHighAvailabilityComponentState(ctx)
}

func (c *networkDeviceCommunicator) GetHighAvailabilityComponentRole(ctx context.Context) (string, error) {
	if !c.HasComponent(component.HighAvailability) {
		return "", tholaerr.NewComponentNotFoundError("no ha component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHighAvailabilityComponentRole(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHighAvailabilityComponentRole(ctx)
}

func (c *networkDeviceCommunicator) GetHighAvailabilityComponentNodes(ctx context.Context) (int, error) {
	if !c.HasComponent(component.HighAvailability) {
		return 0, tholaerr.NewComponentNotFoundError("no ha component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHighAvailabilityComponentNodes(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetHighAvailabilityComponentNodes(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentServicesCurrentlyDown(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentServicesCurrentlyDown(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentServicesCurrentlyDown(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentSystemVersion(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentSystemVersion(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentSystemVersion(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentSIEM(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentSIEM(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentSIEM(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentCpuConsumptionCollection(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentCpuConsumptionCollection(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentCpuConsumptionCollection(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentCpuConsumptionNormalization(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentCpuConsumptionNormalization(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentCpuConsumptionNormalization(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentCpuConsumptionEnrichment(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentCpuConsumptionEnrichment(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentCpuConsumptionEnrichment(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentCpuConsumptionIndexing(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentCpuConsumptionIndexing(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentCpuConsumptionIndexing(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentCpuConsumptionDashboardAlerts(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentCpuConsumptionDashboardAlerts(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentCpuConsumptionDashboardAlerts(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentMemoryConsumptionCollection(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentMemoryConsumptionCollection(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentMemoryConsumptionCollection(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentMemoryConsumptionNormalization(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentMemoryConsumptionNormalization(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentMemoryConsumptionNormalization(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentMemoryConsumptionEnrichment(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentMemoryConsumptionEnrichment(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentMemoryConsumptionEnrichment(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentMemoryConsumptionIndexing(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentMemoryConsumptionIndexing(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentMemoryConsumptionIndexing(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentMemoryConsumptionDashboardAlerts(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentMemoryConsumptionDashboardAlerts(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentMemoryConsumptionDashboardAlerts(ctx)
}

//

func (c *networkDeviceCommunicator) GetSIEMComponentQueueCollection(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentQueueCollection(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentQueueCollection(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentQueueNormalization(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentQueueNormalization(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentQueueNormalization(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentQueueEnrichment(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentQueueEnrichment(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentQueueEnrichment(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentQueueIndexing(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentQueueIndexing(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentQueueIndexing(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentQueueDashboardAlerts(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentQueueDashboardAlerts(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentQueueDashboardAlerts(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentActiveSearchProcesses(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentActiveSearchProcesses(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentActiveSearchProcesses(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentDiskUsageDashboardAlerts(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentDiskUsageDashboardAlerts(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentDiskUsageDashboardAlerts(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentZFSPools(ctx context.Context) ([]device.SIEMComponentZFSPool, error) {
	if !c.HasComponent(component.SIEM) {
		return nil, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentZFSPools(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentZFSPools(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentRepositories(ctx context.Context) ([]device.SIEMComponentRepository, error) {
	if !c.HasComponent(component.SIEM) {
		return nil, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentRepositories(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentRepositories(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerVersion(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerVersion(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerVersion(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerIOWait(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerIOWait(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerIOWait(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerVMSwapiness(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerVMSwapiness(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerVMSwapiness(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerClusterSize(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerClusterSize(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerClusterSize(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerProxyCpuUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerProxyCpuUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerProxyCpuUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerProxyMemoryUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerProxyMemoryUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerProxyMemoryUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerProxyNumberOfAliveConnections(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerProxyNumberOfAliveConnections(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerProxyNumberOfAliveConnections(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerProxyState(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerProxyState(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerProxyState(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerProxyNodesCount(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerProxyNodesCount(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerProxyNodesCount(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerStorageCPUUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerStorageCPUUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerStorageCPUUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerStorageMemoryUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerStorageMemoryUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerStorageMemoryUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentConfiguredCapacity(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentConfiguredCapacity(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentConfiguredCapacity(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerStorageAvailableCapacity(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerStorageAvailableCapacity(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerStorageAvailableCapacity(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerStorageDFSUsed(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerStorageDFSUsed(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerStorageDFSUsed(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerStorageUnderReplicatedBlocks(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerStorageUnderReplicatedBlocks(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerStorageUnderReplicatedBlocks(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerStorageLiveDataNodes(ctx context.Context) (int, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerStorageLiveDataNodes(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerStorageLiveDataNodes(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerAuthenticatorCPUUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerAuthenticatorCPUUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerAuthenticatorCPUUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerAuthenticatorMemoryUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerAuthenticatorMemoryUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerAuthenticatorMemoryUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerAuthenticatorServiceStatus(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerAuthenticatorServiceStatus(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerAuthenticatorServiceStatus(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerAuthenticatorAdminServiceStatus(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerAuthenticatorAdminServiceStatus(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerAuthenticatorAdminServiceStatus(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentFabricServerZFSPools(ctx context.Context) ([]device.SIEMComponentZFSPool, error) {
	if !c.HasComponent(component.SIEM) {
		return nil, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentFabricServerZFSPools(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return nil, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentFabricServerZFSPools(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAPIServerVersion(ctx context.Context) (string, error) {
	if !c.HasComponent(component.SIEM) {
		return "", tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAPIServerVersion(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return "", errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAPIServerVersion(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAPIServerIOWait(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAPIServerIOWait(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAPIServerIOWait(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAPIServerVMSwapiness(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAPIServerVMSwapiness(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAPIServerVMSwapiness(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAPIServerCPUUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAPIServerCPUUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAPIServerCPUUsage(ctx)
}

func (c *networkDeviceCommunicator) GetSIEMComponentAPIServerMemoryUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.SIEM) {
		return 0, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetSIEMComponentAPIServerMemoryUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
			}
		} else {
			return res, nil
		}
	}

	return c.deviceClassCommunicator.GetSIEMComponentAPIServerMemoryUsage(ctx)
}
