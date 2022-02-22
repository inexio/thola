package deviceclass

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/component"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/tholaerr"
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

func (o *deviceClassCommunicator) UpdateConnection(ctx context.Context) error {
	if conn, ok := network.DeviceConnectionFromContext(ctx); ok {
		if conn.SNMP != nil && conn.SNMP.SnmpClient != nil {
			if conn.RawConnectionData.SNMP.MaxRepetitions == nil || *conn.RawConnectionData.SNMP.MaxRepetitions == 0 {
				log.Ctx(ctx).Debug().Uint32("max_repetitions", o.deviceClass.config.snmp.MaxRepetitions).Msg("set snmp max repetitions of device class")
				conn.SNMP.SnmpClient.SetMaxRepetitions(o.deviceClass.config.snmp.MaxRepetitions)
			}

			if conn.SNMP.SnmpClient.GetVersion() != "1" {
				log.Ctx(ctx).Debug().Int("max_oids", o.deviceClass.config.snmp.MaxOids).Msg("set snmp max oids of device class")
				err := conn.SNMP.SnmpClient.SetMaxOIDs(o.deviceClass.config.snmp.MaxOids)
				if err != nil {
					return errors.Wrap(err, "failed to set max oids")
				}
			}
		}
	}
	return nil
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

func (o *deviceClassCommunicator) GetHighAvailabilityComponent(ctx context.Context) (device.HighAvailabilityComponent, error) {
	if !o.HasComponent(component.HighAvailability) {
		return device.HighAvailabilityComponent{}, tholaerr.NewComponentNotFoundError("no ha component available for this device")
	}

	var ha device.HighAvailabilityComponent

	empty := true

	state, err := o.GetHighAvailabilityComponentState(ctx)
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

	role, err := o.GetHighAvailabilityComponentRole(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HighAvailabilityComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		ha.Role = &role
		empty = false
	}

	nodes, err := o.GetHighAvailabilityComponentNodes(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.HighAvailabilityComponent{}, errors.Wrap(err, "error occurred during get high availability nodes")
		}
	} else {
		ha.Nodes = &nodes
		empty = false
	}

	if empty {
		return device.HighAvailabilityComponent{}, tholaerr.NewNotFoundError("no hardware health data available")
	}

	return ha, nil
}

func (o *deviceClassCommunicator) GetSIEMComponent(ctx context.Context) (device.SIEMComponent, error) {
	if !o.HasComponent(component.SIEM) {
		return device.SIEMComponent{}, tholaerr.NewComponentNotFoundError("no siem component available for this device")
	}

	var siem device.SIEMComponent

	empty := true

	lrmpsNormalizer, err := o.GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx)
	if err != nil {
		if !tholaerr.IsNotFoundError(err) && !tholaerr.IsNotImplementedError(err) {
			return device.SIEMComponent{}, errors.Wrap(err, "error occurred during get high availability role")
		}
	} else {
		siem.LastRecordedMessagesPerSecondNormalizer = &lrmpsNormalizer
		empty = false
	}

	if empty {
		return device.SIEMComponent{}, tholaerr.NewNotFoundError("no SIEM data available")
	}

	return siem, nil
}

func (o *deviceClassCommunicator) GetVendor(ctx context.Context) (string, error) {
	if o.identify.properties.vendor == nil {
		log.Ctx(ctx).Debug().Str("property", "vendor").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "vendor").Logger()
	ctx = logger.WithContext(ctx)
	vendor, err := o.identify.properties.vendor.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get vendor")
	}

	return strings.TrimSpace(vendor.String()), nil
}

func (o *deviceClassCommunicator) GetModel(ctx context.Context) (string, error) {
	if o.identify.properties.model == nil {
		log.Ctx(ctx).Debug().Str("property", "model").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "model").Logger()
	ctx = logger.WithContext(ctx)
	model, err := o.identify.properties.model.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get model")
	}

	return strings.TrimSpace(model.String()), nil
}

func (o *deviceClassCommunicator) GetModelSeries(ctx context.Context) (string, error) {
	if o.identify.properties.modelSeries == nil {
		log.Ctx(ctx).Debug().Str("property", "model_series").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "model_series").Logger()
	ctx = logger.WithContext(ctx)
	modelSeries, err := o.identify.properties.modelSeries.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get model_series")
	}

	return strings.TrimSpace(modelSeries.String()), nil
}

func (o *deviceClassCommunicator) GetSerialNumber(ctx context.Context) (string, error) {
	if o.identify.properties.serialNumber == nil {
		log.Ctx(ctx).Debug().Str("property", "serial_number").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "serial_number").Logger()
	ctx = logger.WithContext(ctx)
	serialNumber, err := o.identify.properties.serialNumber.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get serial_number")
	}

	return strings.TrimSpace(serialNumber.String()), nil
}

func (o *deviceClassCommunicator) GetOSVersion(ctx context.Context) (string, error) {
	if o.identify.properties.osVersion == nil {
		log.Ctx(ctx).Debug().Str("property", "osVersion").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "osVersion").Logger()
	ctx = logger.WithContext(ctx)
	version, err := o.identify.properties.osVersion.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get osVersion")
	}

	return strings.TrimSpace(version.String()), nil
}

func (o *deviceClassCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	if o.components.interfaces == nil || o.components.interfaces.properties == nil {
		log.Ctx(ctx).Debug().Str("property", "interfaces").Str("device_class", o.name).Msg("no interface information available")
		return nil, tholaerr.NewNotImplementedError("not implemented")
	}

	interfacesRaw, indices, err := o.components.interfaces.properties.GetProperty(ctx, filter...)
	if err != nil {
		return nil, err
	}

	var interfaces []device.Interface

	err = interfacesRaw.Decode(&interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode raw interfaces into interface structs")
	}

	// normalize interfaces
	for i, interf := range interfaces {
		if interf.IfIndex == nil {
			ifIndex, err := indices[i].UInt64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get ifIndex from SNMP index")
			}
			interfaces[i].IfIndex = &ifIndex
		}
		if interf.IfSpeed != nil && interf.IfHighSpeed != nil && *interf.IfSpeed == math.MaxUint32 {
			ifSpeed := *interf.IfHighSpeed * 1000000
			interfaces[i].IfSpeed = &ifSpeed
		}
	}

	return interfaces, nil
}

func (o *deviceClassCommunicator) GetCountInterfaces(ctx context.Context) (int, error) {
	if o.components.interfaces == nil || o.components.interfaces.count == nil {
		log.Ctx(ctx).Debug().Str("property", "countInterfaces").Str("device_class", o.name).Msg("no interface count information available")
		return 0, tholaerr.NewNotImplementedError("not implemented")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Debug().Msg("snmp client is empty")
		return 0, errors.New("snmp client is empty")
	}

	res, err := o.components.interfaces.count.GetProperty(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get interfaces count")
	}

	if responseInt, err := res.Int(); err == nil {
		return responseInt, nil
	}

	return 0, errors.New("could not parse response to int")
}

func (o *deviceClassCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]device.CPU, error) {
	if o.components.cpu == nil || o.components.cpu.properties == nil {
		log.Ctx(ctx).Debug().Str("property", "CPUComponentCPULoad").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "CPUComponentCPULoad").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.cpu.properties.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return nil, errors.Wrap(err, "failed to get CPUComponentCPULoad")
	}
	var cpus []device.CPU
	err = res.Decode(&cpus)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode group properties into CPUs array")
	}
	return cpus, nil
}

func (o *deviceClassCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) ([]device.MemoryPool, error) {
	if o.components.memory == nil || o.components.memory.usage == nil {
		log.Ctx(ctx).Debug().Str("property", "MemoryComponentMemoryUsage").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "MemoryComponentMemoryUsage").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.memory.usage.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return nil, errors.Wrap(err, "failed to get MemoryComponentMemoryUsage")
	}

	var pools []device.MemoryPool
	err = res.Decode(&pools)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode group properties into CPUs array")
	}
	return pools, nil
}

func (o *deviceClassCommunicator) GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error) {
	if o.components.disk == nil || o.components.disk.properties == nil {
		log.Ctx(ctx).Debug().Str("groupProperty", "DiskComponentStorages").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "DiskComponentStorages").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.disk.properties.GetProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var storages []device.DiskComponentStorage
	err = mapstructure.WeakDecode(res, &storages)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into storage struct")
	}
	return storages, nil
}

func (o *deviceClassCommunicator) GetUPSComponentAlarmLowVoltageDisconnect(ctx context.Context) (int, error) {
	if o.components.ups == nil || o.components.ups.alarmLowVoltageDisconnect == nil {
		log.Ctx(ctx).Debug().Str("property", "UPSComponentAlarmLowVoltageDisconnect").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentAlarmAlarmLowVoltageDisconnect").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.alarmLowVoltageDisconnect.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentBatteryAmperage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryAmperage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryAmperage.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentBatteryCapacity").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryCapacity").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryCapacity.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentBatteryCurrent").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryCurrent").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryCurrent.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentBatteryRemainingTime").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryRemainingTime").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryRemainingTime.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentBatteryTemperature").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryTemperature").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryTemperature.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentBatteryVoltage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentBatteryVoltage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.batteryVoltage.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentCurrentLoad").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentCurrentLoad").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.currentLoad.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentMainsVoltageApplied").Str("device_class", o.name).Msg("no detection information available")
		return false, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentMainsVoltageApplied").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.mainsVoltageApplied.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentRectifierCurrent").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentRectifierCurrent").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.rectifierCurrent.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "UPSComponentSystemVoltage").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "UPSComponentSystemVoltage").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.ups.systemVoltage.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("groupProperty", "SBCComponentAgents").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "SBCComponentAgents").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.sbc.agents.GetProperty(ctx)
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
		log.Ctx(ctx).Debug().Str("groupProperty", "SBCComponentRealms").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "SBCComponentRealms").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.sbc.realms.GetProperty(ctx)
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentGlobalCallPerSecond").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentGlobalCallPerSecond").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.globalCallPerSecond.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentGlobalConcurrentSessions").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentGlobalConcurrentSessions").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.globalConcurrentSessions.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentActiveLocalContacts").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentActiveLocalContacts").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.activeLocalContacts.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentTranscodingCapacity").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentTranscodingCapacity").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.transcodingCapacity.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentLicenseCapacity").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentLicenseCapacity").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.licenseCapacity.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentSystemRedundancy").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentSystemRedundancy").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.systemRedundancy.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "SBCComponentSystemHealthScore").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "SBCComponentSystemHealthScore").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.sbc.systemHealthScore.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "ServerComponentProcs").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "ServerComponentProcs").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.server.procs.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
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
		log.Ctx(ctx).Debug().Str("property", "ServerComponentUsers").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "ServerComponentUsers").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.server.users.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get ServerComponentUsers")
	}
	r, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}
	return r, nil
}

func (o *deviceClassCommunicator) GetHardwareHealthComponentEnvironmentMonitorState(ctx context.Context) (device.HardwareHealthComponentState, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.environmentMonitorState == nil {
		log.Ctx(ctx).Debug().Str("property", "HardwareHealthComponentEnvironmentMonitorState").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "HardwareHealthComponentEnvironmentMonitorState").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.hardwareHealth.environmentMonitorState.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get HardwareHealthComponentEnvironmentMonitorState")
	}

	state := device.HardwareHealthComponentState(res.String())
	if _, err := state.GetInt(); err != nil {
		return "", fmt.Errorf("read out invalid hardware health component state '%s'", state)
	}
	return state, nil
}

func (o *deviceClassCommunicator) GetHardwareHealthComponentFans(ctx context.Context) ([]device.HardwareHealthComponentFan, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.fans == nil {
		log.Ctx(ctx).Debug().Str("groupProperty", "HardwareHealthComponentFans").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "HardwareHealthComponentFans").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.hardwareHealth.fans.GetProperty(ctx)
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
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.powerSupply == nil {
		log.Ctx(ctx).Debug().Str("groupProperty", "HardwareHealthComponentPowerSupply").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "HardwareHealthComponentPowerSupply").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.hardwareHealth.powerSupply.GetProperty(ctx)
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

func (o *deviceClassCommunicator) GetHardwareHealthComponentTemperature(ctx context.Context) ([]device.HardwareHealthComponentTemperature, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.temperature == nil {
		log.Ctx(ctx).Debug().Str("groupProperty", "HardwareHealthComponentFans").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "HardwareHealthComponentTemperature").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.hardwareHealth.temperature.GetProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var temperatures []device.HardwareHealthComponentTemperature
	err = res.Decode(&temperatures)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into temperature struct")
	}
	return temperatures, nil
}

func (o *deviceClassCommunicator) GetHardwareHealthComponentVoltage(ctx context.Context) ([]device.HardwareHealthComponentVoltage, error) {
	if o.components.hardwareHealth == nil || o.components.hardwareHealth.voltage == nil {
		log.Ctx(ctx).Debug().Str("groupProperty", "HardwareHealthComponentVoltage").Str("device_class", o.name).Msg("no detection information available")
		return nil, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("groupProperty", "HardwareHealthComponentTemperature").Logger()
	ctx = logger.WithContext(ctx)
	res, _, err := o.components.hardwareHealth.voltage.GetProperty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get property")
	}
	var voltage []device.HardwareHealthComponentVoltage
	err = res.Decode(&voltage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode property into voltage struct")
	}
	return voltage, nil
}

func (o *deviceClassCommunicator) GetHighAvailabilityComponentState(ctx context.Context) (device.HighAvailabilityComponentState, error) {
	if o.components.highAvailability == nil || o.components.highAvailability.state == nil {
		log.Ctx(ctx).Debug().Str("property", "HighAvailabilityComponentState").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "HighAvailabilityComponentState").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.highAvailability.state.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get HighAvailabilityComponentState")
	}

	state := device.HighAvailabilityComponentState(res.String())
	if _, err := state.GetInt(); err != nil {
		return "", fmt.Errorf("read out invalid highAvailability component state '%s'", state)
	}
	return state, nil
}

func (o *deviceClassCommunicator) GetHighAvailabilityComponentRole(ctx context.Context) (string, error) {
	if o.components.highAvailability == nil || o.components.highAvailability.role == nil {
		log.Ctx(ctx).Debug().Str("property", "HighAvailabilityComponentRole").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "HighAvailabilityComponentRole").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.highAvailability.role.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get HighAvailabilityComponentRole")
	}

	return res.String(), nil
}

func (o *deviceClassCommunicator) GetHighAvailabilityComponentNodes(ctx context.Context) (int, error) {
	if o.components.highAvailability == nil || o.components.highAvailability.nodes == nil {
		log.Ctx(ctx).Debug().Str("property", "HighAvailabilityComponentNodes").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}
	logger := log.Ctx(ctx).With().Str("property", "HighAvailabilityComponentNodes").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.highAvailability.nodes.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get HighAvailabilityComponentNodes")
	}

	v, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentLastRecordedMessagesPerSecondNormalizer(ctx context.Context) (int, error) {
	if o.components.siem == nil || o.components.siem.lastRecordedMessagesPerSecondNormalizer == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentLastRecordedMessagesPerSecondNormalizer").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentLastRecordedMessagesPerSecondNormalizer").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.lastRecordedMessagesPerSecondNormalizer.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SIEMComponentLastRecordedMessagesPerSecondNormalizer")
	}

	v, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentAverageMessagesPerSecondLast5minNormalizer(ctx context.Context) (int, error) {
	if o.components.siem == nil || o.components.siem.averageMessagesPerSecondLast5minNormalizer == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentAverageMessagesPerSecondLast5minNormalizer").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentAverageMessagesPerSecondLast5minNormalizer").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.averageMessagesPerSecondLast5minNormalizer.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SIEMComponentAverageMessagesPerSecondLast5minNormalizer")
	}

	v, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentLastRecordedMessagesPerSecondStoreHandler(ctx context.Context) (int, error) {
	if o.components.siem == nil || o.components.siem.lastRecordedMessagesPerSecondStoreHandler == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentLastRecordedMessagesPerSecondStoreHandler").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentLastRecordedMessagesPerSecondStoreHandler").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.lastRecordedMessagesPerSecondStoreHandler.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SIEMComponentLastRecordedMessagesPerSecondStoreHandler")
	}

	v, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentAverageMessagesPerSecondLast5minStoreHandler(ctx context.Context) (int, error) {
	if o.components.siem == nil || o.components.siem.averageMessagesPerSecondLast5minStoreHandler == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentAverageMessagesPerSecondLast5minStoreHandler").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentAverageMessagesPerSecondLast5minStoreHandler").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.averageMessagesPerSecondLast5minStoreHandler.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SIEMComponentAverageMessagesPerSecondLast5minStoreHandler")
	}

	v, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentServicesCurrentlyDown(ctx context.Context) (int, error) {
	if o.components.siem == nil || o.components.siem.servicesCurrentlyDown == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentServicesCurrentlyDown").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentServicesCurrentlyDown").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.servicesCurrentlyDown.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SIEMComponentServicesCurrentlyDown")
	}

	v, err := res.Int()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to int", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentSystemVersion(ctx context.Context) (string, error) {
	if o.components.siem == nil || o.components.siem.systemVersion == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentSystemVersion").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentSystemVersion").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.systemVersion.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get SIEMComponentSystemVersion")
	}

	return res.String(), nil
}

func (o *deviceClassCommunicator) GetSIEMComponentSIEM(ctx context.Context) (string, error) {
	if o.components.siem == nil || o.components.siem.systemVersion == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentSIEM").Str("device_class", o.name).Msg("no detection information available")
		return "", tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentSIEM").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.siem.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return "", errors.Wrap(err, "failed to get SIEMComponentSIEM")
	}

	return res.String(), nil
}

func (o *deviceClassCommunicator) GetSIEMComponentCpuConsumptionCollection(ctx context.Context) (float64, error) {
	if o.components.siem == nil || o.components.siem.cpuConsumptionCollection == nil {
		log.Ctx(ctx).Debug().Str("property", "SIEMComponentCpuConsumptionCollection").Str("device_class", o.name).Msg("no detection information available")
		return 0, tholaerr.NewNotImplementedError("no detection information available")
	}

	logger := log.Ctx(ctx).With().Str("property", "SIEMComponentCpuConsumptionCollection").Logger()
	ctx = logger.WithContext(ctx)
	res, err := o.components.siem.cpuConsumptionCollection.GetProperty(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get property")
		return 0, errors.Wrap(err, "failed to get SIEMComponentCpuConsumptionCollection")
	}

	v, err := res.Float64()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert value '%s' to float64", res.String())
	}

	return v, nil
}

func (o *deviceClassCommunicator) GetSIEMComponentCpuConsumptionNormalization(ctx context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (o *deviceClassCommunicator) GetSIEMComponentCpuConsumptionEnrichment(ctx context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (o *deviceClassCommunicator) GetSIEMComponentCpuConsumptionIndexing(ctx context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}

func (o *deviceClassCommunicator) GetSIEMComponentCpuConsumptionDashboardAlerts(ctx context.Context) (float64, error) {
	return 0, tholaerr.NewNotImplementedError("function is not implemented for this communicator")
}
