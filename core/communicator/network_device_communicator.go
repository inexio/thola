package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/utility"
	"github.com/pkg/errors"
	"math"
)

// NetworkDeviceCommunicator represents a communicator for a device.
type NetworkDeviceCommunicator interface {

	// GetDeviceClass returns the device class of a network device.
	GetDeviceClass() string

	// GetAvailableComponents returns the components available for a network device.
	GetAvailableComponents() []string

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

	availableCommunicatorFunctions
}

type availableCommunicatorFunctions interface {

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

	// GetIfTable returns the ifTable of a device.
	// This only contains the standard ifTable values.
	GetIfTable(ctx context.Context) ([]device.Interface, error)

	// GetInterfaces returns the interfaces of a device.
	// This includes special interface values.
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

type networkDeviceCommunicator struct {
	*relatedNetworkDeviceCommunicators
	codeCommunicator        availableCommunicatorFunctions
	deviceClassCommunicator *deviceClassCommunicator
}

// GetDeviceClass returns the OS.
func (c *networkDeviceCommunicator) GetDeviceClass() string {
	return c.deviceClassCommunicator.getName()
}

// GetAvailableComponents returns the available Components for the device.
func (c *networkDeviceCommunicator) GetAvailableComponents() []string {
	var res []string
	components := c.deviceClassCommunicator.getAvailableComponents()
	for k, v := range components {
		if v {
			component, err := k.toString()
			if err != nil {
				continue
			}
			res = append(res, component)
		}
	}
	return res
}

func (c *networkDeviceCommunicator) executeWithRecursion(fClass, fCom, fSub adapterFunc, args ...interface{}) (interface{}, error) {
	var err1, err2, err3 error

	value, err1 := fClass(args...)
	if err1 == nil {
		return value, nil
	} else if !tholaerr.IsNotImplementedError(err1) {
		return nil, err1
	}

	if fCom != nil {
		value, err2 = fCom(args...)
		if err2 == nil {
			return value, nil
		} else if !tholaerr.IsNotImplementedError(err2) {
			return nil, err2
		}
	} else {
		err2 = tholaerr.NewNotImplementedError("no communicator available")
	}

	if fSub != nil {
		value, err3 = fSub(args...)
		if err3 == nil {
			return value, err3
		} else if !tholaerr.IsNotFoundError(err3) {
			return nil, err3
		}
	} else {
		err3 = tholaerr.NewNotImplementedError("no parent communicator with implementation available")
	}

	if tholaerr.IsNotImplementedError(err1) && tholaerr.IsNotImplementedError(err2) && tholaerr.IsNotImplementedError(err3) {
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	return nil, tholaerr.NewNotFoundError("failed to get information through any device class")
}

//
// Identify properties functions are defined here.
//

func (c *networkDeviceCommunicator) GetIdentifyProperties(ctx context.Context) (device.Properties, error) {
	dev := device.Device{
		Class:      c.head.GetDeviceClass(),
		Properties: device.Properties{},
	}

	vendor, err := c.head.GetVendor(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get vendor")
		}
	} else {
		dev.Properties.Vendor = &vendor
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	model, err := c.head.GetModel(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model")
		}
	} else {
		dev.Properties.Model = &model
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	modelSeries, err := c.head.GetModelSeries(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model series")
		}
	} else {
		dev.Properties.ModelSeries = &modelSeries
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	serialNumber, err := c.head.GetSerialNumber(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get serial number")
		}
	} else {
		dev.Properties.SerialNumber = &serialNumber
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	osVersion, err := c.head.GetOSVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get os version")
		}
	} else {
		dev.Properties.OSVersion = &osVersion
	}

	return dev.Properties, nil
}

func (c *networkDeviceCommunicator) GetVendor(ctx context.Context) (string, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getVendor
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getVendor), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getVendor), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return "", err
	}
	s := res.(string)
	if s == "" {
		return "", tholaerr.NewNotFoundError("empty string returned")
	}
	return s, err
}

func (c *networkDeviceCommunicator) GetModel(ctx context.Context) (string, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getModel
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getModel), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getModel), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return "", err
	}
	s := res.(string)
	if s == "" {
		return "", tholaerr.NewNotFoundError("empty string returned")
	}
	return s, err
}

func (c *networkDeviceCommunicator) GetModelSeries(ctx context.Context) (string, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getModelSeries
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getModelSeries), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getModelSeries), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return "", err
	}
	s := res.(string)
	if s == "" {
		return "", tholaerr.NewNotFoundError("empty string returned")
	}
	return s, err
}

func (c *networkDeviceCommunicator) GetSerialNumber(ctx context.Context) (string, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSerialNumber
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSerialNumber), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSerialNumber), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return "", err
	}
	s := res.(string)
	if s == "" {
		return "", tholaerr.NewNotFoundError("empty string returned")
	}
	return s, err
}

func (c *networkDeviceCommunicator) GetOSVersion(ctx context.Context) (string, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getOSVersion
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getOSVersion), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getOSVersion), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return "", err
	}
	s := res.(string)
	if s == "" {
		return "", tholaerr.NewNotFoundError("empty string returned")
	}
	return s, err
}

//
// Interface component functions are defined here.
//

func (c *networkDeviceCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	if c.isHead() && !c.deviceClassCommunicator.hasAvailableComponent(interfacesComponent) {
		return []device.Interface{}, tholaerr.NewComponentNotFoundError("no interfaces component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getInterfaces
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getInterfaces), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getInterfaces), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return []device.Interface{}, err
	}
	if c.isHead() {
		res = normalizeInterfaces(res.([]device.Interface))
	}
	return res.([]device.Interface), err
}

func (c *networkDeviceCommunicator) GetIfTable(ctx context.Context) ([]device.Interface, error) {
	if c.isHead() && !c.deviceClassCommunicator.hasAvailableComponent(interfacesComponent) {
		return []device.Interface{}, tholaerr.NewComponentNotFoundError("no interfaces component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getIfTable
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getIfTable), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getIfTable), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return []device.Interface{}, err
	}
	return res.([]device.Interface), err
}

func (c *networkDeviceCommunicator) GetCountInterfaces(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getCountInterfaces
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getCountInterfaces), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getCountInterfaces), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func normalizeInterfaces(interfaces []device.Interface) []device.Interface {
	for i, interf := range interfaces {
		if interf.IfSpeed != nil && interf.IfHighSpeed != nil && *interf.IfSpeed == math.MaxUint32 {
			ifSpeed := *interf.IfHighSpeed * 1000000
			interfaces[i].IfSpeed = &ifSpeed
		}
	}
	return interfaces
}

//
// CPU component functions are defined here.
//

func (c *networkDeviceCommunicator) GetCPUComponent(ctx context.Context) (device.CPUComponent, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(cpuComponent) {
		return device.CPUComponent{}, tholaerr.NewComponentNotFoundError("no cpu component available for this device")
	}

	var cpu device.CPUComponent
	empty := true

	cpuLoad, err := c.head.GetCPUComponentCPULoad(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.CPUComponent{}, errors.Wrap(err, "error occurred during get cpu load")
		}
	} else {
		cpu.Load = cpuLoad
		empty = false
	}

	cpuTemp, err := c.head.GetCPUComponentCPUTemperature(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.CPUComponent{}, errors.Wrap(err, "error occurred during get cpu temperature")
		}
	} else {
		cpu.Temperature = cpuTemp
		empty = false
	}

	if empty {
		return device.CPUComponent{}, tholaerr.NewNotFoundError("no cpu data available")
	}
	return cpu, nil
}

func (c *networkDeviceCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]float64, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(cpuComponent) {
		return nil, tholaerr.NewComponentNotFoundError("no cpu component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getCPULoad
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getCPULoad), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getCPULoad), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]float64), err
}

func (c *networkDeviceCommunicator) GetCPUComponentCPUTemperature(ctx context.Context) ([]float64, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(cpuComponent) {
		return nil, tholaerr.NewComponentNotFoundError("no cpu component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getCPUTemperature
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getCPUTemperature), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getCPUTemperature), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]float64), err
}

//
// Memory component functions are defined here.
//

func (c *networkDeviceCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) (float64, error) {
	if c.isHead() && !c.deviceClassCommunicator.hasAvailableComponent(memoryComponent) {
		return 0, tholaerr.NewComponentNotFoundError("no memory component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getMemoryUsage
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getMemoryUsage), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getMemoryUsage), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

//
// Disk component functions are defined here.
//

func (c *networkDeviceCommunicator) GetDiskComponent(ctx context.Context) (device.DiskComponent, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(diskComponent) {
		return device.DiskComponent{}, tholaerr.NewComponentNotFoundError("no disk component available for this device")
	}

	var disk device.DiskComponent

	empty := true

	storages, err := c.head.GetDiskComponentStorages(ctx)
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

func (c *networkDeviceCommunicator) GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getDiskComponentStorages
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getDiskComponentStorages), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getDiskComponentStorages), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]device.DiskComponentStorage), err
}

//
// UPS component functions are defined here.
//

func (c *networkDeviceCommunicator) GetUPSComponent(ctx context.Context) (device.UPSComponent, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(upsComponent) {
		return device.UPSComponent{}, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	var ups device.UPSComponent
	empty := true

	alarmLowVoltage, err := c.head.GetUPSComponentAlarmLowVoltageDisconnect(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get alarm")
		}
	} else {
		ups.AlarmLowVoltageDisconnect = &alarmLowVoltage
		empty = false
	}

	batteryAmperage, err := c.head.GetUPSComponentBatteryAmperage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery amperage")
		}
	} else {
		ups.BatteryAmperage = &batteryAmperage
		empty = false
	}

	batteryCapacity, err := c.head.GetUPSComponentBatteryCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryCapacity = &batteryCapacity
		empty = false
	}

	batteryCurrent, err := c.head.GetUPSComponentBatteryCurrent(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryCurrent = &batteryCurrent
		empty = false
	}

	batteryRemainingTime, err := c.head.GetUPSComponentBatteryRemainingTime(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryRemainingTime = &batteryRemainingTime
		empty = false
	}

	batteryTemperature, err := c.head.GetUPSComponentBatteryTemperature(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery temperature")
		}
	} else {
		ups.BatteryTemperature = &batteryTemperature
		empty = false
	}

	batteryVoltage, err := c.head.GetUPSComponentBatteryVoltage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery voltage")
		}
	} else {
		ups.BatteryVoltage = &batteryVoltage
		empty = false
	}

	currentLoad, err := c.head.GetUPSComponentCurrentLoad(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get current load")
		}
	} else {
		ups.CurrentLoad = &currentLoad
		empty = false
	}

	mainsVoltageApplied, err := c.head.GetUPSComponentMainsVoltageApplied(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.MainsVoltageApplied = &mainsVoltageApplied
		empty = false
	}

	rectifierCurrent, err := c.head.GetUPSComponentRectifierCurrent(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.RectifierCurrent = &rectifierCurrent
		empty = false
	}

	systemVoltage, err := c.head.GetUPSComponentSystemVoltage(ctx)
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

func (c *networkDeviceCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentAlarmLowVoltageDisconnect
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentAlarmLowVoltageDisconnect), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentAlarmLowVoltageDisconnect), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryAmperage(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentBatteryAmperage
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentBatteryAmperage), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentBatteryAmperage), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryCapacity(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentBatteryCapacity
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentBatteryCapacity), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentBatteryCapacity), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryCurrent(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentBatteryCurrent
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentBatteryCurrent), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentBatteryCurrent), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryRemainingTime(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentBatteryRemainingTime
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentBatteryRemainingTime), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentBatteryRemainingTime), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryTemperature(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentBatteryTemperature
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentBatteryTemperature), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentBatteryTemperature), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentBatteryVoltage(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentBatteryVoltage
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentBatteryVoltage), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentBatteryVoltage), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentCurrentLoad(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentCurrentLoad
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentCurrentLoad), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentCurrentLoad), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentMainsVoltageApplied
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentMainsVoltageApplied), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentMainsVoltageApplied), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return false, err
	}
	return res.(bool), err
}

func (c *networkDeviceCommunicator) GetUPSComponentRectifierCurrent(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentRectifierCurrent
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentRectifierCurrent), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentRectifierCurrent), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

func (c *networkDeviceCommunicator) GetUPSComponentSystemVoltage(ctx context.Context) (float64, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getUPSComponentSystemVoltage
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getUPSComponentSystemVoltage), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getUPSComponentSystemVoltage), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(float64), err
}

//
// Server component functions are defined here.
//

func (c *networkDeviceCommunicator) GetServerComponent(ctx context.Context) (device.ServerComponent, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(serverComponent) {
		return device.ServerComponent{}, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}

	var server device.ServerComponent

	empty := true

	procs, err := c.head.GetServerComponentProcs(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.ServerComponent{}, errors.Wrap(err, "error occurred during get server component procs")
		}
	} else {
		server.Procs = &procs
		empty = false
	}

	users, err := c.head.GetServerComponentUsers(ctx)
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

func (c *networkDeviceCommunicator) GetServerComponentProcs(ctx context.Context) (int, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(serverComponent) {
		return 0, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getServerProcs
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getServerProcs), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getServerProcs), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetServerComponentUsers(ctx context.Context) (int, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(serverComponent) {
		return 0, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getServerUsers
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getServerUsers), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getServerUsers), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

//
// SBC component functions are defined here.
//

func (c *networkDeviceCommunicator) GetSBCComponent(ctx context.Context) (device.SBCComponent, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(sbcComponent) {
		return device.SBCComponent{}, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	var sbc device.SBCComponent

	empty := true

	agents, err := c.head.GetSBCComponentAgents(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component agents")
		}
	} else {
		sbc.Agents = agents
		empty = false
	}

	realms, err := c.head.GetSBCComponentRealms(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component realms")
		}
	} else {
		sbc.Realms = realms
		empty = false
	}

	globalCPS, err := c.head.GetSBCComponentGlobalCallPerSecond(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component sbc global call per second")
		}
	} else {
		sbc.GlobalCallPerSecond = &globalCPS
		empty = false
	}

	globalConcurrentSessions, err := c.head.GetSBCComponentGlobalConcurrentSessions(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc global concurrent sessions")
		}
	} else {
		sbc.GlobalConcurrentSessions = &globalConcurrentSessions
		empty = false
	}

	activeLocalContacts, err := c.head.GetSBCComponentActiveLocalContacts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get active local contacts")
		}
	} else {
		sbc.ActiveLocalContacts = &activeLocalContacts
		empty = false
	}

	transcodingCapacity, err := c.head.GetSBCComponentTranscodingCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get transcoding capacity")
		}
	} else {
		sbc.TranscodingCapacity = &transcodingCapacity
		empty = false
	}

	licenseCapacity, err := c.head.GetSBCComponentLicenseCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get license capacity")
		}
	} else {
		sbc.LicenseCapacity = &licenseCapacity
		empty = false
	}

	systemRedundancy, err := c.head.GetSBCComponentSystemRedundancy(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get system redundancy")
		}
	} else {
		sbc.SystemRedundancy = &systemRedundancy
		empty = false
	}

	systemHealthScore, err := c.head.GetSBCComponentSystemHealthScore(ctx)
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

func (c *networkDeviceCommunicator) GetSBCComponentAgents(ctx context.Context) ([]device.SBCComponentAgent, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentAgents
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentAgents), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentAgents), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]device.SBCComponentAgent), err
}

func (c *networkDeviceCommunicator) GetSBCComponentRealms(ctx context.Context) ([]device.SBCComponentRealm, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentRealms
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentRealms), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentRealms), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]device.SBCComponentRealm), err
}

func (c *networkDeviceCommunicator) GetSBCComponentGlobalCallPerSecond(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentGlobalCallPerSecond
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentGlobalCallPerSecond), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentGlobalCallPerSecond), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetSBCComponentGlobalConcurrentSessions(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentGlobalConcurrentSessions
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentGlobalConcurrentSessions), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentGlobalConcurrentSessions), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetSBCComponentActiveLocalContacts(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentActiveLocalContacts
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentActiveLocalContacts), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentActiveLocalContacts), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetSBCComponentTranscodingCapacity(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentTranscodingCapacity
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentTranscodingCapacity), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentTranscodingCapacity), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetSBCComponentLicenseCapacity(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentLicenseCapacity
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentLicenseCapacity), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentLicenseCapacity), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetSBCComponentSystemRedundancy(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentSystemRedundancy
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentSystemRedundancy), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentSystemRedundancy), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetSBCComponentSystemHealthScore(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getSBCComponentSystemHealthScore
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getSBCComponentSystemHealthScore), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getSBCComponentSystemHealthScore), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

//
// Hardware health component functions are defined here.
//

func (c *networkDeviceCommunicator) GetHardwareHealthComponent(ctx context.Context) (device.HardwareHealthComponent, error) {
	if !c.deviceClassCommunicator.hasAvailableComponent(hardwareHealthComponent) {
		return device.HardwareHealthComponent{}, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	var hardwareHealth device.HardwareHealthComponent

	empty := true

	state, err := c.head.GetHardwareHealthComponentEnvironmentMonitorState(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get environment monitor states")
		}
	} else {
		hardwareHealth.EnvironmentMonitorState = &state
		empty = false
	}

	fans, err := c.head.GetHardwareHealthComponentFans(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get fans")
		}
	} else {
		hardwareHealth.Fans = fans
		empty = false
	}

	powerSupply, err := c.head.GetHardwareHealthComponentPowerSupply(ctx)
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

func (c *networkDeviceCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (int, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getHardwareHealthComponentEnvironmentMonitorState
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getHardwareHealthComponentEnvironmentMonitorState), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getHardwareHealthComponentEnvironmentMonitorState), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return 0, err
	}
	return res.(int), err
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentFans(ctx context.Context) ([]device.HardwareHealthComponentFan, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getHardwareHealthComponentFans
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getHardwareHealthComponentFans), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getHardwareHealthComponentFans), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]device.HardwareHealthComponentFan), err
}

func (c *networkDeviceCommunicator) GetHardwareHealthComponentPowerSupply(ctx context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	fClass := newCommunicatorAdapter(c.deviceClassCommunicator).getHardwareHealthComponentPowerSupply
	fCom := utility.IfThenElse(c.codeCommunicator != nil, adapterFunc(newCommunicatorAdapter(c.codeCommunicator).getHardwareHealthComponentPowerSupply), emptyAdapterFunc).(adapterFunc)
	fSub := utility.IfThenElse(c.sub != nil, adapterFunc(newCommunicatorAdapter(c.sub).getHardwareHealthComponentPowerSupply), emptyAdapterFunc).(adapterFunc)
	res, err := c.executeWithRecursion(fClass, fCom, fSub, ctx)
	if err != nil {
		return nil, err
	}
	return res.([]device.HardwareHealthComponentPowerSupply), err
}

func (c *networkDeviceCommunicator) isHead() bool {
	return c.head.GetDeviceClass() == c.GetDeviceClass()
}
