package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/inexio/thola/core/value"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

type deviceClassCommunicator struct {
	baseCommunicator
	*deviceClass
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
	if o.components.interfaces == nil || (o.components.interfaces.IfTable == nil && o.components.interfaces.Types == nil) {
		log.Ctx(ctx).Trace().Str("property", "interfaces").Str("device_class", o.name).Msg("no interface information available")
		return nil, tholaerr.NewNotImplementedError("not implemented")
	}

	networkInterfaces, err := o.head.GetIfTable(ctx)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get ifTable")
		return nil, errors.Wrap(err, "failed to get ifTable")
	}

	for _, typeDef := range o.components.interfaces.Types {
		specialInterfacesRaw, err := o.getValuesBySNMPWalk(ctx, typeDef.Values)
		if err != nil {
			return nil, err
		}

		for i, networkInterface := range networkInterfaces {
			if specialValues, ok := specialInterfacesRaw[fmt.Sprint(*networkInterface.IfIndex)]; ok {
				err := mapstructure.WeakDecode(specialValues, &networkInterfaces[i])
				if err != nil {
					log.Ctx(ctx).Trace().Err(err).Msg("can't parse oid values into Interface struct")
					return nil, errors.Wrap(err, "can't parse oid values into Interface struct")
				}
			}
		}
	}

	return networkInterfaces, nil
}

func (o *deviceClassCommunicator) GetIfTable(ctx context.Context) ([]device.Interface, error) {
	if o.components.interfaces == nil || o.components.interfaces.IfTable == nil {
		log.Ctx(ctx).Trace().Str("property", "ifTable").Str("device_class", o.name).Msg("no interface information available")
		return nil, tholaerr.NewNotImplementedError("not implemented")
	}

	networkInterfacesRaw, err := o.components.interfaces.IfTable.getProperty(ctx)
	if err != nil {
		return nil, err
	}

	var networkInterfaces []device.Interface

	for _, oidValue := range networkInterfacesRaw {
		var networkInterface device.Interface
		err := mapstructure.WeakDecode(oidValue, &networkInterface)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msg("can't parse oid values into Interface struct")
			return nil, errors.Wrap(err, "can't parse oid values into Interface struct")
		}
		networkInterfaces = append(networkInterfaces, networkInterface)
	}

	return networkInterfaces, nil
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

	interfaces, err := o.head.GetInterfaces(ctx)
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

func (o *deviceClassCommunicator) getValuesBySNMPWalk(ctx context.Context, oids deviceClassOIDs) (map[string]map[string]interface{}, error) {
	networkInterfaces := make(map[string]map[string]interface{})

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		log.Ctx(ctx).Trace().Str("property", "interface").Msg("snmp client is empty")
		return nil, errors.New("snmp client is empty")
	}

	for name, oid := range oids {
		snmpResponse, err := con.SNMP.SnmpClient.SNMPWalk(ctx, string(oid.OID))
		if err != nil {
			if tholaerr.IsNotFoundError(err) {
				log.Ctx(ctx).Trace().Err(err).Msgf("oid %s (%s) not found on device", oid.OID, name)
				continue
			}
			log.Ctx(ctx).Trace().Err(err).Msg("failed to get oid value of interface")
			return nil, errors.Wrap(err, "failed to get oid value")
		}

		for _, response := range snmpResponse {
			res, err := response.GetValueBySNMPGetConfiguration(oid.SNMPGetConfiguration)
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("couldn't get value from response response")
				return nil, errors.Wrap(err, "couldn't get value from response response")
			}
			if res != "" {
				resNormalized, err := oid.operators.apply(ctx, value.New(res))
				if err != nil {
					log.Ctx(ctx).Trace().Err(err).Msg("response couldn't be normalized")
					return nil, errors.Wrap(err, "response couldn't be normalized")
				}
				oid := strings.Split(response.GetOID(), ".")
				ifIndex := oid[len(oid)-1]
				if _, ok := networkInterfaces[ifIndex]; !ok {
					networkInterfaces[ifIndex] = make(map[string]interface{})
				}
				networkInterfaces[ifIndex][name] = resNormalized
			}
		}
	}

	return networkInterfaces, nil
}
