package deviceclass

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/communicator/component"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"math"
	"strings"
)

type deviceClassCommunicator struct {
	*deviceClass
}

func (o *deviceClassCommunicator) GetIdentifier() string {
	return o.getName()
}

func (o *deviceClassCommunicator) GetAvailableComponents() []string {
	var res []string
	components := o.getAvailableComponents()
	for k, v := range components {
		if v {
			comp, err := k.ToString()
			if err != nil {
				continue
			}
			res = append(res, comp)
		}
	}
	return res
}

func (o *deviceClassCommunicator) HasComponent(component component.Component) bool {
	haha := o.getAvailableComponents()
	if v, ok := haha[component]; ok && v {
		return true
	}
	return false
}

func (o *deviceClassCommunicator) Match(ctx context.Context) (bool, error) {
	return o.matchDevice(ctx)
}

func (o *deviceClassCommunicator) GetIdentifyProperties(ctx context.Context) (device.Properties, error) {
	dev := device.Device{
		Class:      o.GetIdentifier(),
		Properties: device.Properties{},
	}

	vendor, err := o.GetVendor(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get vendor")
		}
	} else {
		dev.Properties.Vendor = &vendor
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	model, err := o.GetModel(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model")
		}
	} else {
		dev.Properties.Model = &model
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	modelSeries, err := o.GetModelSeries(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get model series")
		}
	} else {
		dev.Properties.ModelSeries = &modelSeries
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	serialNumber, err := o.GetSerialNumber(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get serial number")
		}
	} else {
		dev.Properties.SerialNumber = &serialNumber
		ctx = device.NewContextWithDeviceProperties(ctx, dev)
	}

	osVersion, err := o.GetOSVersion(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.Properties{}, errors.Wrap(err, "error occurred during get os version")
		}
	} else {
		dev.Properties.OSVersion = &osVersion
	}

	return dev.Properties, nil
}

func (o *deviceClassCommunicator) GetCPUComponent(ctx context.Context) (device.CPUComponent, error) {
	if !o.HasComponent(component.CPU) {
		return device.CPUComponent{}, tholaerr.NewComponentNotFoundError("no cpu component available for this device")
	}

	var cpu device.CPUComponent
	empty := true

	cpuLoad, err := o.GetCPUComponentCPULoad(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.CPUComponent{}, errors.Wrap(err, "error occurred during get cpu load")
		}
	} else {
		cpu.Load = cpuLoad
		empty = false
	}

	cpuTemp, err := o.GetCPUComponentCPUTemperature(ctx)
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

func (o *deviceClassCommunicator) GetDiskComponent(ctx context.Context) (device.DiskComponent, error) {
	if !o.HasComponent(component.Disk) {
		return device.DiskComponent{}, tholaerr.NewComponentNotFoundError("no disk component available for this device")
	}

	var disk device.DiskComponent

	empty := true

	storages, err := o.GetDiskComponentStorages(ctx)
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

func (o *deviceClassCommunicator) GetUPSComponent(ctx context.Context) (device.UPSComponent, error) {
	if !o.HasComponent(component.UPS) {
		return device.UPSComponent{}, tholaerr.NewComponentNotFoundError("no ups component available for this device")
	}

	var ups device.UPSComponent
	empty := true

	alarmLowVoltage, err := o.GetUPSComponentAlarmLowVoltageDisconnect(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get alarm")
		}
	} else {
		ups.AlarmLowVoltageDisconnect = &alarmLowVoltage
		empty = false
	}

	batteryAmperage, err := o.GetUPSComponentBatteryAmperage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery amperage")
		}
	} else {
		ups.BatteryAmperage = &batteryAmperage
		empty = false
	}

	batteryCapacity, err := o.GetUPSComponentBatteryCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryCapacity = &batteryCapacity
		empty = false
	}

	batteryCurrent, err := o.GetUPSComponentBatteryCurrent(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryCurrent = &batteryCurrent
		empty = false
	}

	batteryRemainingTime, err := o.GetUPSComponentBatteryRemainingTime(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery capacity")
		}
	} else {
		ups.BatteryRemainingTime = &batteryRemainingTime
		empty = false
	}

	batteryTemperature, err := o.GetUPSComponentBatteryTemperature(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery temperature")
		}
	} else {
		ups.BatteryTemperature = &batteryTemperature
		empty = false
	}

	batteryVoltage, err := o.GetUPSComponentBatteryVoltage(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get battery voltage")
		}
	} else {
		ups.BatteryVoltage = &batteryVoltage
		empty = false
	}

	currentLoad, err := o.GetUPSComponentCurrentLoad(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get current load")
		}
	} else {
		ups.CurrentLoad = &currentLoad
		empty = false
	}

	mainsVoltageApplied, err := o.GetUPSComponentMainsVoltageApplied(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.MainsVoltageApplied = &mainsVoltageApplied
		empty = false
	}

	rectifierCurrent, err := o.GetUPSComponentRectifierCurrent(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.UPSComponent{}, errors.Wrap(err, "error occurred during get mains voltage applied")
		}
	} else {
		ups.RectifierCurrent = &rectifierCurrent
		empty = false
	}

	systemVoltage, err := o.GetUPSComponentSystemVoltage(ctx)
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

func (o *deviceClassCommunicator) GetServerComponent(ctx context.Context) (device.ServerComponent, error) {
	if !o.HasComponent(component.Server) {
		return device.ServerComponent{}, tholaerr.NewComponentNotFoundError("no server component available for this device")
	}

	var server device.ServerComponent

	empty := true

	procs, err := o.GetServerComponentProcs(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.ServerComponent{}, errors.Wrap(err, "error occurred during get server component procs")
		}
	} else {
		server.Procs = &procs
		empty = false
	}

	users, err := o.GetServerComponentUsers(ctx)
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

func (o *deviceClassCommunicator) GetSBCComponent(ctx context.Context) (device.SBCComponent, error) {
	if !o.HasComponent(component.SBC) {
		return device.SBCComponent{}, tholaerr.NewComponentNotFoundError("no sbc component available for this device")
	}

	var sbc device.SBCComponent

	empty := true

	agents, err := o.GetSBCComponentAgents(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component agents")
		}
	} else {
		sbc.Agents = agents
		empty = false
	}

	realms, err := o.GetSBCComponentRealms(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component realms")
		}
	} else {
		sbc.Realms = realms
		empty = false
	}

	globalCPS, err := o.GetSBCComponentGlobalCallPerSecond(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc component sbc global call per second")
		}
	} else {
		sbc.GlobalCallPerSecond = &globalCPS
		empty = false
	}

	globalConcurrentSessions, err := o.GetSBCComponentGlobalConcurrentSessions(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get sbc global concurrent sessions")
		}
	} else {
		sbc.GlobalConcurrentSessions = &globalConcurrentSessions
		empty = false
	}

	activeLocalContacts, err := o.GetSBCComponentActiveLocalContacts(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get active local contacts")
		}
	} else {
		sbc.ActiveLocalContacts = &activeLocalContacts
		empty = false
	}

	transcodingCapacity, err := o.GetSBCComponentTranscodingCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get transcoding capacity")
		}
	} else {
		sbc.TranscodingCapacity = &transcodingCapacity
		empty = false
	}

	licenseCapacity, err := o.GetSBCComponentLicenseCapacity(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get license capacity")
		}
	} else {
		sbc.LicenseCapacity = &licenseCapacity
		empty = false
	}

	systemRedundancy, err := o.GetSBCComponentSystemRedundancy(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SBCComponent{}, errors.Wrap(err, "error occurred during get system redundancy")
		}
	} else {
		sbc.SystemRedundancy = &systemRedundancy
		empty = false
	}

	systemHealthScore, err := o.GetSBCComponentSystemHealthScore(ctx)
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

func (o *deviceClassCommunicator) GetHardwareHealthComponent(ctx context.Context) (device.HardwareHealthComponent, error) {
	if !o.HasComponent(component.HardwareHealth) {
		return device.HardwareHealthComponent{}, tholaerr.NewComponentNotFoundError("no hardware health component available for this device")
	}

	var hardwareHealth device.HardwareHealthComponent

	empty := true

	state, err := o.GetHardwareHealthComponentEnvironmentMonitorState(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get environment monitor states")
		}
	} else {
		hardwareHealth.EnvironmentMonitorState = &state
		empty = false
	}

	fans, err := o.GetHardwareHealthComponentFans(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HardwareHealthComponent{}, errors.Wrap(err, "error occurred during get fans")
		}
	} else {
		hardwareHealth.Fans = fans
		empty = false
	}

	powerSupply, err := o.GetHardwareHealthComponentPowerSupply(ctx)
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

func (o *deviceClassCommunicator) GetVendor(ctx context.Context) (string, error) {
	if o.identify.properties.vendor == nil {
		log.Ctx(ctx).Trace().Str("property", "vendor").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "vendor").Logger()
	ctx = logger.WithContext(ctx)
	vendor, err := o.identify.properties.vendor.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get vendor")
	}

	return strings.TrimSpace(vendor.String()), nil
}

func (o *deviceClassCommunicator) GetModel(ctx context.Context) (string, error) {
	if o.identify.properties.model == nil {
		log.Ctx(ctx).Trace().Str("property", "model").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "model").Logger()
	ctx = logger.WithContext(ctx)
	model, err := o.identify.properties.model.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get model")
	}

	return strings.TrimSpace(model.String()), nil
}

func (o *deviceClassCommunicator) GetModelSeries(ctx context.Context) (string, error) {
	if o.identify.properties.modelSeries == nil {
		log.Ctx(ctx).Trace().Str("property", "model_series").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "model_series").Logger()
	ctx = logger.WithContext(ctx)
	modelSeries, err := o.identify.properties.modelSeries.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get model_series")
	}

	return strings.TrimSpace(modelSeries.String()), nil
}

func (o *deviceClassCommunicator) GetSerialNumber(ctx context.Context) (string, error) {
	if o.identify.properties.serialNumber == nil {
		log.Ctx(ctx).Trace().Str("property", "serial_number").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "serial_number").Logger()
	ctx = logger.WithContext(ctx)
	serialNumber, err := o.identify.properties.serialNumber.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get serial_number")
	}

	return strings.TrimSpace(serialNumber.String()), nil
}

func (o *deviceClassCommunicator) GetOSVersion(ctx context.Context) (string, error) {
	if o.identify.properties.osVersion == nil {
		log.Ctx(ctx).Trace().Str("property", "osVersion").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "osVersion").Logger()
	ctx = logger.WithContext(ctx)
	version, err := o.identify.properties.osVersion.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get osVersion")
	}

	return strings.TrimSpace(version.String()), nil
}

func (o *deviceClassCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	if o.components.interfaces == nil || o.components.interfaces.Values == nil {
		log.Ctx(ctx).Trace().Str("property", "interfaces").Str("device_class", o.name).Msg("no interface information available")
		return nil, tholaerr.NewNotImplementedError("not implemented")
	}

	interfacesRaw, err := o.components.interfaces.Values.getProperty(ctx)
	if err != nil {
		return nil, err
	}

	var interfaces []device.Interface

	err = interfacesRaw.Decode(&interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode raw interfaces into interface structs")
	}

	for i, interf := range interfaces {
		if interf.IfSpeed != nil && interf.IfHighSpeed != nil && *interf.IfSpeed == math.MaxUint32 {
			ifSpeed := *interf.IfHighSpeed * 1000000
			interfaces[i].IfSpeed = &ifSpeed
		}
	}

	return interfaces, nil
}

func (o *deviceClassCommunicator) GetCountInterfaces(ctx context.Context) (int, error) {
	if o.components.interfaces == nil || o.components.interfaces.Count == "" {
		log.Ctx(ctx).Trace().Str("property", "countInterfaces").Str("device_class", o.name).Msg("no interface count information available")
		return 0, tholaerr.NewNotImplementedError("not implemented")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Trace().Msg("snmp client is empty")
		return 0, errors.New("snmp client is empty")
	}

	oid := o.components.interfaces.Count

	snmpResponse, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)

	if err == nil {
		response, err := snmpResponse[0].GetValue()
		if err == nil {
			if responseInt, ok := response.(int); ok {
				return responseInt, nil
			}
			err := fmt.Errorf("could not parse response to int, response has type %T", response)
			log.Ctx(ctx).Trace().Err(err).Msgf("could not parse response to int, response has type %T", response)
			return 0, err
		}
		log.Ctx(ctx).Trace().Err(err).Msg("response is empty")
		return 0, errors.Wrap(err, "response is empty")
	}

	interfaces, err := o.GetInterfaces(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to read out interfaces")
		return 0, errors.Wrap(err, "failed to read out interfaces")
	}

	return len(interfaces), nil
}

func (o *deviceClassCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]float64, error) {
	if o.components.cpu == nil || o.components.cpu.load == nil {
		log.Ctx(ctx).Trace().Str("property", "CPUComponentCPULoad").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "CPUComponentCPULoad").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.cpu.load.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return nil, errors.Wrap(err, "failed to get CPUComponentCPULoad")
	}
	r, err := res.Float64()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}
	return []float64{r}, nil
}

func (o *deviceClassCommunicator) GetCPUComponentCPUTemperature(ctx context.Context) ([]float64, error) {
	if o.components.cpu == nil || o.components.cpu.temperature == nil {
		log.Ctx(ctx).Trace().Str("property", "CPUComponentCPUTemperature").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "CPUComponentCPUTemperature").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.cpu.temperature.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return nil, errors.Wrap(err, "failed to get CPUComponentCPUTemperature")
	}
	r, err := res.Float64()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert value '%s' to float64", res.String())
	}
	return []float64{r}, nil
}

func (o *deviceClassCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) (float64, error) {
	if o.components.memory == nil || o.components.memory.usage == nil {
		log.Ctx(ctx).Trace().Str("property", "MemoryComponentMemoryUsage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "MemoryComponentMemoryUsage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.memory.usage.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get MemoryComponentMemoryUsage")
	}
	r, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to float64", res.String())
	}
	return r, nil
}

func (o *deviceClassCommunicator) GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error) {
	if o.components.disk == nil || o.components.disk.storages == nil {
		log.Ctx(ctx).Trace().Str("groupProperty", "DiskComponentStorages").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "DiskComponentStorages").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.disk.storages.getProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var storages []device.DiskComponentStorage
	err = mapstructure.WeakDecode(res, &storages)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into storage struct")
	}
	// ignore non-physical storage types
	var filtered []device.DiskComponentStorage
	for _, storage := range storages {
		if *storage.Type != "Other" && *storage.Type != "RAM" && *storage.Type != "Virtual Memory" {
			filtered = append(filtered, storage)
		}
	}
	return filtered, nil
}

func (o *deviceClassCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(ctx context.Context) (int, error) {
	if o.components.ups == nil || o.components.ups.alarmLowVoltageDisconnect == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentAlarmLowVoltageDisconnect").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentAlarmAlarmLowVoltageDisconnect").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.alarmLowVoltageDisconnect.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentAlarmAlarmLowVoltageDisconnect")
	}
	r, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}
	return r, nil
}

func (o *deviceClassCommunicator) GetUPSComponentBatteryAmperage(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.batteryAmperage == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentBatteryAmperage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryAmperage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryAmperage.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentBatteryAmperage")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentBatteryCapacity(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.batteryCapacity == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentBatteryCapacity").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryCapacity").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryCapacity.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentBatteryCapacity")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentBatteryCurrent(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.batteryCurrent == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentBatteryCurrent").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryCurrent").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryCurrent.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentBatteryCurrent")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentBatteryRemainingTime(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.batteryRemainingTime == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentBatteryRemainingTime").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryRemainingTime").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryRemainingTime.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentBatteryRemainingTime")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentBatteryTemperature(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.batteryTemperature == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentBatteryTemperature").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryTemperature").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryTemperature.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentBatteryTemperature")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentBatteryVoltage(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.batteryVoltage == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentBatteryVoltage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryVoltage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryVoltage.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentBatteryVoltage")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentCurrentLoad(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.currentLoad == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentCurrentLoad").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentCurrentLoad").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.currentLoad.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentCurrentLoad")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error) {
	if o.components.ups == nil || o.components.ups.mainsVoltageApplied == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentMainsVoltageApplied").Str("device_class", o.name).Msg("no detection information available")
		return false, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentMainsVoltageApplied").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.mainsVoltageApplied.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return false, errors.Wrap(err, "failed to get UPSComponentMainsVoltageApplied")
	}
	r, err := res.Bool()
	if err != nil {
		return false, errors.Wrapf(err, "failed to parse value '%s' to bool", res.String())
	}
	return r, nil
}

func (o *deviceClassCommunicator) GetUPSComponentRectifierCurrent(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.rectifierCurrent == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentRectifierCurrent").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentRectifierCurrent").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.rectifierCurrent.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentRectifierCurrent")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetUPSComponentSystemVoltage(ctx context.Context) (float64, error) {
	if o.components.ups == nil || o.components.ups.systemVoltage == nil {
		log.Ctx(ctx).Trace().Str("property", "UPSComponentSystemVoltage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentSystemVoltage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.systemVoltage.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get UPSComponentSystemVoltage")
	}
	result, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to float64", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentAgents(ctx context.Context) ([]device.SBCComponentAgent, error) {
	if o.components.sbc == nil || o.components.sbc.agents == nil {
		log.Ctx(ctx).Trace().Str("groupProperty", "SBCComponentAgents").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "SBCComponentAgents").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.agents.getProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var agents []device.SBCComponentAgent
	err = mapstructure.WeakDecode(res, &agents)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into agent struct")
	}
	return agents, nil
}

func (o *deviceClassCommunicator) GetSBCComponentRealms(ctx context.Context) ([]device.SBCComponentRealm, error) {
	if o.components.sbc == nil || o.components.sbc.realms == nil {
		log.Ctx(ctx).Trace().Str("groupProperty", "SBCComponentRealms").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "SBCComponentRealms").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.realms.getProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var realms []device.SBCComponentRealm
	err = mapstructure.WeakDecode(res, &realms)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into realms struct")
	}
	return realms, nil
}

func (o *deviceClassCommunicator) GetSBCComponentGlobalCallPerSecond(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.globalCallPerSecond == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentGlobalCallPerSecond").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentGlobalCallPerSecond").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.globalCallPerSecond.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentGlobalCallPerSecond")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentGlobalConcurrentSessions(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.globalConcurrentSessions == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentGlobalConcurrentSessions").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentGlobalConcurrentSessions").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.globalConcurrentSessions.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentGlobalConcurrentSessions")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentActiveLocalContacts(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.activeLocalContacts == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentActiveLocalContacts").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentActiveLocalContacts").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.activeLocalContacts.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentActiveLocalContacts")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentTranscodingCapacity(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.transcodingCapacity == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentTranscodingCapacity").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentTranscodingCapacity").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.transcodingCapacity.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentTranscodingCapacity")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentLicenseCapacity(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.licenseCapacity == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentLicenseCapacity").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentLicenseCapacity").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.licenseCapacity.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentLicenseCapacity")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentSystemRedundancy(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.systemRedundancy == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentSystemRedundancy").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentSystemRedundancy").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.systemRedundancy.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentSystemRedundancy")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetSBCComponentSystemHealthScore(ctx context.Context) (int, error) {
	if o.components.sbc == nil || o.components.sbc.systemHealthScore == nil {
		log.Ctx(ctx).Trace().Str("property", "SBCComponentSystemHealthScore").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentSystemHealthScore").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.systemHealthScore.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SBCComponentSystemHealthScore")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetServerComponentProcs(ctx context.Context) (int, error) {
	if o.components.server == nil || o.components.server.procs == nil {
		log.Ctx(ctx).Trace().Str("property", "ServerComponentProcs").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "ServerComponentProcs").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.server.procs.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get ServerComponentProcs")
	}
	r, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}
	return r, nil
}

func (o *deviceClassCommunicator) GetServerComponentUsers(ctx context.Context) (int, error) {
	if o.components.server == nil || o.components.server.users == nil {
		log.Ctx(ctx).Trace().Str("property", "ServerComponentUsers").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "ServerComponentUsers").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.server.users.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get ServerComponentUsers")
	}
	r, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}
	return r, nil
}

func (o *deviceClassCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (int, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.environmentMonitorState == nil {
		log.Ctx(ctx).Trace().Str("property", "HardwareHealthComponentEnvironmentMonitorState").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "HardwareHealthComponentEnvironmentMonitorState").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.hardwareHealth.environmentMonitorState.getProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get HardwareHealthComponentEnvironmentMonitorState")
	}
	result, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert result '%v' to int", res)
	}
	return result, nil
}

func (o *deviceClassCommunicator) GetHardwareHealthComponentFans(ctx context.Context) ([]device.HardwareHealthComponentFan, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.fans == nil {
		log.Ctx(ctx).Trace().Str("groupProperty", "HardwareHealthComponentFans").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "HardwareHealthComponentFans").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.hardwareHealth.fans.getProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var fans []device.HardwareHealthComponentFan
	err = mapstructure.WeakDecode(res, &fans)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into fan struct")
	}
	return fans, nil
}

func (o *deviceClassCommunicator) GetHardwareHealthComponentPowerSupply(ctx context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.fans == nil {
		log.Ctx(ctx).Trace().Str("groupProperty", "HardwareHealthComponentPowerSupply").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "HardwareHealthComponentPowerSupply").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.hardwareHealth.powerSupply.getProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var powerSupply []device.HardwareHealthComponentPowerSupply
	err = mapstructure.WeakDecode(res, &powerSupply)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into power supply struct")
	}
	return powerSupply, nil
}
