package codecommunicator

import (
	"context"
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
