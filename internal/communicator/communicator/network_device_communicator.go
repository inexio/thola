package communicator

import (
	"context"
	"github.com/inexio/thola/internal/communicator/component"
	"github.com/inexio/thola/internal/communicator/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
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
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get sbc component sbc global call per second")
		}
	} else {
		hardwareHealth.PowerSupply = powerSupply
		empty = false
	}

	if empty {
		return device.HardwareHealthComponent{}, tholaerr.NewNotFoundError("no sbc data available")
	}

	return hardwareHealth, nil
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
		var interfaces []device.Interface
		interfaces, err = c.GetInterfaces(ctx)
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

func (c *networkDeviceCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) (float64, error) {
	if !c.HasComponent(component.Memory) {
		return 0, tholaerr.NewComponentNotFoundError("no memory component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetMemoryComponentMemoryUsage(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
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

func (c *networkDeviceCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (int, error) {
	if !c.HasComponent(component.HardwareHealth) {
		return 0, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	if c.codeCommunicator != nil {
		res, err := c.codeCommunicator.GetHardwareHealthComponentEnvironmentMonitorState(ctx)
		if err != nil {
			if !tholaerr.IsNotImplementedError(err) {
				return 0, errors.Wrap(err, "error in code communicator")
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
