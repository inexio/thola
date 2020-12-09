package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/utility"
	"github.com/pkg/errors"
)

// NetworkDeviceCommunicator represents a communicator for a device
type NetworkDeviceCommunicator interface {
	GetDeviceClass() string
	GetAvailableComponents() []string
	GetIdentifyProperties(ctx context.Context) (device.Properties, error)
	GetCPUComponent(ctx context.Context) (device.CPUComponent, error)
	GetUPSComponent(ctx context.Context) (device.UPSComponent, error)
	availableCommunicatorFunctions
}

type availableCommunicatorFunctions interface {
	GetVendor(ctx context.Context) (string, error)
	GetModel(ctx context.Context) (string, error)
	GetModelSeries(ctx context.Context) (string, error)
	GetSerialNumber(ctx context.Context) (string, error)
	GetOSVersion(ctx context.Context) (string, error)
	GetIfTable(ctx context.Context) ([]device.Interface, error)
	GetInterfaces(ctx context.Context) ([]device.Interface, error)
	GetCountInterfaces(ctx context.Context) (int, error)
	availableCPUCommunicatorFunctions
	availableUPSCommunicatorFunctions
}

type availableCPUCommunicatorFunctions interface {
	GetCPUComponentCPULoad(ctx context.Context) ([]float64, error)
	GetCPUComponentCPUTemperature(ctx context.Context) ([]float64, error)
}

type availableUPSCommunicatorFunctions interface {
	GetUPSComponentAlarmLowVoltageDisconnect(ctx context.Context) (int, error)
	GetUPSComponentBatteryAmperage(ctx context.Context) (float64, error)
	GetUPSComponentBatteryCapacity(ctx context.Context) (float64, error)
	GetUPSComponentBatteryCurrent(ctx context.Context) (float64, error)
	GetUPSComponentBatteryRemainingTime(ctx context.Context) (float64, error)
	GetUPSComponentBatteryTemperature(ctx context.Context) (float64, error)
	GetUPSComponentBatteryVoltage(ctx context.Context) (float64, error)
	GetUPSComponentCurrentLoad(ctx context.Context) (float64, error)
	GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error)
	GetUPSComponentRectifierCurrent(ctx context.Context) (float64, error)
	GetUPSComponentSystemVoltage(ctx context.Context) (float64, error)
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
		return nil, errors.Wrap(err1, "an unexpected error occurred while trying to get value through device class")
	}

	if fCom != nil {
		value, err2 = fCom(args...)
		if err2 == nil {
			return value, nil
		} else if !tholaerr.IsNotImplementedError(err2) {
			return nil, errors.Wrap(err2, "an unexpected error occurred while trying to get value through communicator")
		}
	} else {
		err2 = tholaerr.NewNotImplementedError("no communicator available")
	}

	if fSub != nil {
		value, err3 = fSub(args...)
		if err3 == nil {
			return value, err3
		} else if !tholaerr.IsNotFoundError(err3) {
			return nil, errors.Wrap(err3, "an unexpected error occurred while trying to get value in parent communicator")
		}
	} else {
		err3 = tholaerr.NewNotImplementedError("no parent communicator with implementation available")
	}

	if tholaerr.IsNotImplementedError(err1) && tholaerr.IsNotImplementedError(err2) && tholaerr.IsNotImplementedError(err3) {
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	return nil, tholaerr.NewNotFoundError("failed to get value")
}

func (c *networkDeviceCommunicator) GetIdentifyProperties(ctx context.Context) (device.Properties, error) {
	var properties device.Properties

	vendor, err := c.head.GetVendor(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get vendor")
		}
	} else {
		properties.Vendor = &vendor
	}

	model, err := c.head.GetModel(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model")
		}
	} else {
		properties.Model = &model
	}

	modelSeries, err := c.head.GetModelSeries(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model series")
		}
	} else {
		properties.ModelSeries = &modelSeries
	}

	serialNumber, err := c.head.GetSerialNumber(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get serial number")
		}
	} else {
		properties.SerialNumber = &serialNumber
	}

	osVersion, err := c.head.GetOSVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get os version")
		}
	} else {
		properties.OSVersion = &osVersion
	}

	return properties, nil
}

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
	if c.isRoot() {
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

func (c *networkDeviceCommunicator) isHead() bool {
	return c.head.GetDeviceClass() == c.GetDeviceClass()
}

func (c *networkDeviceCommunicator) isRoot() bool {
	return c.GetDeviceClass() == "generic"
}

func normalizeInterfaces(interfaces []device.Interface) []device.Interface {
	for i := range interfaces {
		if interfaces[i].IfSpeed != nil {
			var speed uint64
			if interfaces[i].IfHighSpeed != nil && *interfaces[i].IfSpeed == 4294967295 {
				speed = *interfaces[i].IfHighSpeed * 1000000
			} else {
				speed = *interfaces[i].IfSpeed
			}
			//if radio interface
			if interfaces[i].LevelIn != nil {
				speed *= 1000
			}
			interfaces[i].IfSpeed = &speed
		}
	}
	return interfaces
}
