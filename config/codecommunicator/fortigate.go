package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/value"
	"github.com/pkg/errors"
	"regexp"
)

type fortigateCommunicator struct {
	codeCommunicator
}

type fortigateSensorData struct {
	Name  *string
	Value value.Value
	State *device.HardwareHealthComponentState
}

func (c *fortigateCommunicator) GetHardwareHealthComponentFans(ctx context.Context) ([]device.HardwareHealthComponentFan, error) {
	regex, err := regexp.Compile(`Fan\s`)
	if err != nil {
		return nil, errors.New("invalid regular expression")
	}

	sensors, err := c.getHardwareHealthComponentReadOutSensors(ctx, regex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out fan sensors")
	}

	var fans []device.HardwareHealthComponentFan
	for _, sensor := range sensors {
		var fan device.HardwareHealthComponentFan
		if sensor.Name != nil {
			fan.Description = sensor.Name
		}
		if sensor.State != nil {
			fan.State = sensor.State
		}
		fans = append(fans, fan)
	}

	return fans, nil
}

func (c *fortigateCommunicator) GetHardwareHealthComponentTemperature(ctx context.Context) ([]device.HardwareHealthComponentTemperature, error) {
	regex, err := regexp.Compile(`Temp|LM75|((TD|TR)\d+)|(DTS\d+)`)
	if err != nil {
		return nil, errors.New("invalid regular expression")
	}

	sensors, err := c.getHardwareHealthComponentReadOutSensors(ctx, regex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out fan sensors")
	}

	var temps []device.HardwareHealthComponentTemperature
	for _, sensor := range sensors {
		var temp device.HardwareHealthComponentTemperature
		if sensor.Name != nil {
			temp.Description = sensor.Name
		}
		if sensor.State != nil {
			temp.State = sensor.State
		}
		if sensor.Value != nil {
			fl, err := sensor.Value.Float64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse temperature as float64")
			}
			temp.Temperature = &fl
		}

		temps = append(temps, temp)
	}

	return temps, nil
}

func (c *fortigateCommunicator) GetHardwareHealthComponentVoltage(ctx context.Context) ([]device.HardwareHealthComponentVoltage, error) {
	regex, err := regexp.Compile(`(VOUT)|(VIN)|(VCC)|(P\d+V\d+)|(_\d+V\d+_)|(DDR)|(VCORE)|(DVDD)`)
	if err != nil {
		return nil, errors.New("invalid regular expression")
	}
	sensors, err := c.getHardwareHealthComponentReadOutSensors(ctx, regex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out voltage sensors")
	}

	var voltage []device.HardwareHealthComponentVoltage
	for _, sensor := range sensors {
		var vol device.HardwareHealthComponentVoltage
		if sensor.Name != nil {
			vol.Description = sensor.Name
		}
		if sensor.State != nil {
			vol.State = sensor.State
		}
		if sensor.Value != nil {
			fl, err := sensor.Value.Float64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse voltage as float64")
			}
			vol.Voltage = &fl
		}

		voltage = append(voltage, vol)
	}

	return voltage, nil
}

func (c *fortigateCommunicator) GetHardwareHealthComponentPowerSupply(ctx context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	regex, err := regexp.Compile(`PS.*Status`)
	if err != nil {
		return nil, errors.New("invalid regular expression")
	}

	sensors, err := c.getHardwareHealthComponentReadOutSensors(ctx, regex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out power supply sensors")
	}

	var powerSupply []device.HardwareHealthComponentPowerSupply
	for _, sensor := range sensors {
		var ps device.HardwareHealthComponentPowerSupply
		if sensor.Name != nil {
			ps.Description = sensor.Name
		}
		if sensor.State != nil {
			ps.State = sensor.State
		}

		powerSupply = append(powerSupply, ps)
	}

	return powerSupply, nil
}

func (c *fortigateCommunicator) getHardwareHealthComponentReadOutSensors(ctx context.Context, regex *regexp.Regexp) ([]fortigateSensorData, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	var sensors []fortigateSensorData

	sensorNameOID := network.OID("1.3.6.1.4.1.12356.101.4.3.2.1.2")
	valueOID := network.OID("1.3.6.1.4.1.12356.101.4.3.2.1.3")
	alarmOID := network.OID("1.3.6.1.4.1.12356.101.4.3.2.1.4")

	sensorResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, sensorNameOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out sensors oid")
	}

	for _, sensorResult := range sensorResults {
		v, err := sensorResult.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value of oid response")
		}

		sensorName := v.String()
		if regex != nil && !regex.MatchString(sensorName) {
			continue
		}

		var sensor fortigateSensorData
		sensor.Name = &sensorName

		oid := sensorResult.GetOID()
		subTree, err := oid.GetIndexAfterOID(sensorNameOID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get subtree of oid %s", sensorResult.GetOID())
		}

		// get value
		resValue, err := con.SNMP.SnmpClient.SNMPGet(ctx, valueOID.AddIndex(subTree))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read out value of sensor %s", sensorName)
		}
		if len(resValue) > 0 {
			v, err := resValue[0].GetValue()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get value for sensor %s", sensorName)
			}
			sensor.Value = v
		}

		// get alarm
		resAlarm, err := con.SNMP.SnmpClient.SNMPGet(ctx, alarmOID.AddIndex(subTree))
		// alarm is not mandatory and does not exist for every sensor
		if err == nil && len(resAlarm) > 0 {
			alarm, err := resAlarm[0].GetValue()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get alarm for sensor %s", sensorName)
			}

			switch alarm.String() {
			case "0":
				state := device.HardwareHealthComponentStateNormal
				sensor.State = &state
			case "1":
				state := device.HardwareHealthComponentStateCritical
				sensor.State = &state
			default:
				state := device.HardwareHealthComponentStateUnknown
				sensor.State = &state
			}
		}

		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

func (c *fortigateCommunicator) GetHighAvailabilityComponentState(ctx context.Context) (device.HighAvailabilityComponentState, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return "", errors.New("no device connection available")
	}

	// check if ha mode is standalone
	modeRes, err := con.SNMP.SnmpClient.SNMPGet(ctx, "1.3.6.1.4.1.12356.101.13.1.1.0") // fgHaSystemMode
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability mode")
	}

	if len(modeRes) < 1 {
		return "", errors.New("failed to read out high-availability mode")
	}

	mode, err := modeRes[0].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability mode")
	}

	if mode.String() == "1" {
		return device.HighAvailabilityComponentStateStandalone, nil
	}

	// if ha mode != standalone, read out sync state
	idx, err := c.getHighAvailabilityIndex(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to find device in high-availability table")
	}

	oid := network.OID("1.3.6.1.4.1.12356.101.13.2.1.1.12") // fgHaStatsSyncStatus

	stateRes, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid.AddIndex(idx))
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability sync state")
	}

	if len(stateRes) < 1 {
		return "", errors.New("failed to read out high-availability mode")
	}

	state, err := stateRes[0].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability sync state")
	}

	var res device.HighAvailabilityComponentState

	switch s := state.String(); s {
	case "0":
		res = device.HighAvailabilityComponentStateUnsynchronized
	case "1":
		res = device.HighAvailabilityComponentStateSynchronized
	default:
		return "", fmt.Errorf("unknown sync state '%s'")
	}

	return res, nil
}

func (c *fortigateCommunicator) GetHighAvailabilityComponentRole(ctx context.Context) (string, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return "", errors.New("no device connection available")
	}

	state, err := c.GetHighAvailabilityComponentState(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability state")
	}

	if state == device.HighAvailabilityComponentStateStandalone {
		return "", errors.New("device is not in high-availability mode (state = standalone)")
	}

	idx, err := c.getHighAvailabilityIndex(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to find device in high-availability table")
	}

	snmpRes, err := con.SNMP.SnmpClient.SNMPGet(ctx, network.OID("1.3.6.1.4.1.12356.101.13.2.1.1.2").AddIndex(idx), network.OID("1.3.6.1.4.1.12356.101.13.2.1.1.16").AddIndex(idx))
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high availability serial numbers")
	}

	if len(snmpRes) != 2 {
		return "", errors.Wrap(err, "failed to read out high availability serial numbers")
	}

	serial, err := snmpRes[0].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability serial number")
	}

	masterSerial, err := snmpRes[1].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability master serial number")
	}

	if serial.IsEmpty() && masterSerial.IsEmpty() {
		return "", errors.New("cannot map serial number to master serial number because one of them is empty")
	}

	if serial.String() == masterSerial.String() {
		return "master", nil
	}
	return "slave", nil
}

func (c *fortigateCommunicator) GetHighAvailabilityComponentNodes(ctx context.Context) (int, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return 0, errors.New("no device connection available")
	}

	state, err := c.GetHighAvailabilityComponentState(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to read out high-availability state")
	}

	if state == device.HighAvailabilityComponentStateStandalone {
		return 0, errors.New("device is not in high-availability mode (state = standalone)")
	}

	snmpRes, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.12356.101.13.2.1.1.1")
	if err != nil {
		return 0, errors.Wrap(err, "failed to read out high availability serial numbers")
	}

	return len(snmpRes), nil
}

func (c *fortigateCommunicator) getHighAvailabilityIndex(ctx context.Context) (string, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return "", errors.New("no device connection available")
	}

	serial, err := c.deviceClass.GetSerialNumber(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to read out serial number")
	}

	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.12356.101.13.2.1.1.2")
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high availability serial numbers")
	}

	for _, r := range res {
		val, err := r.GetValue()
		if err != nil {
			return "", errors.Wrap(err, "failed to get high availability serial number")
		}

		if val.String() == serial {
			return r.GetOID().GetIndex(), nil
		}
	}

	return "", errors.New("failed to get ha index")
}
