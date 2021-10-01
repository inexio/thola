package codecommunicator

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

type iosCommunicator struct {
	codeCommunicator
}

// GetCPUComponentCPULoad returns the cpu load of ios devices.
func (c *iosCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]device.CPU, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}
	var cpus []device.CPU

	cpuLoad5minDeprecated, err1 := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5")
	cpuLoad5min, err2 := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.8")
	if err1 != nil && err2 != nil {
		return nil, errors.New("snmpwalks failed")
	}

	indices := make(map[string]int)

	// save cpus load result for cpuLoad5min
	for _, cpuLoadResponse := range cpuLoad5min {
		cpu, err := c.getCPUBySNMPResponse(cpuLoadResponse)
		if err != nil {
			return nil, err
		}
		cpus = append(cpus, cpu)
		indices[cpuLoadResponse.GetOID().GetIndex()] = len(cpus) - 1 //current entry
	}

	// check deprecated cpu load oid. if one of the entries does not already exist in the cpu arr, add it
	for _, cpuLoadResponseDeprecated := range cpuLoad5minDeprecated {
		idx := cpuLoadResponseDeprecated.GetOID().GetIndex()

		if _, ok := indices[idx]; ok {
			continue
		}

		cpu, err := c.getCPUBySNMPResponse(cpuLoadResponseDeprecated)
		if err != nil {
			return nil, err
		}
		cpus = append(cpus, cpu)
		indices[cpuLoadResponseDeprecated.GetOID().GetIndex()] = len(cpus) - 1 //current entry
	}

	// read out physical indices for cpus
	physicalIndicesResult, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2")
	if err != nil {
		// cannot determine cpu physical indices, return cpu loads without labels
		return cpus, nil
	}

	for _, physicalIndexResult := range physicalIndicesResult {
		idx := physicalIndexResult.GetOID().GetIndex()
		cpuIndex, ok := indices[idx]
		if !ok {
			continue
		}

		physicalIndex, err := physicalIndexResult.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get physical index as string")
		}

		// 0 == physical entry not supported
		if physicalIndex.String() == "0" {
			continue
		}

		physicalNameResponse, err := con.SNMP.SnmpClient.SNMPGet(ctx, network.OID("1.3.6.1.2.1.47.1.1.1.1.7."+physicalIndex.String()))
		if err != nil {
			// cannot get physical name, continue
			continue
		}

		physicalName, err := physicalNameResponse[0].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "physical name is not a string")
		}
		physicalNameString := physicalName.String()

		cpus[cpuIndex].Label = &physicalNameString
	}

	return cpus, nil
}

func (c *iosCommunicator) getCPUBySNMPResponse(res network.SNMPResponse) (device.CPU, error) {
	val, err := res.GetValue()
	if err != nil {
		return device.CPU{}, errors.Wrap(err, "failed to get cpu load value")
	}
	valFloat, err := val.Float64()
	if err != nil {
		return device.CPU{}, errors.Wrap(err, "cpu load is not a float value")
	}
	return device.CPU{
		Label: nil,
		Load:  &valFloat,
	}, nil
}

// GetMemoryComponentMemoryUsage returns the memory usage of ios devices.
func (c *iosCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) ([]device.MemoryPool, error) {
	// first try cisco enhanced mempool mib, if it fails try old mempool mib
	pools, err := c.getMemoryComponentMemoryUsage(ctx,
		".1.3.6.1.4.1.9.9.221.1.1.1.1.3",
		".1.3.6.1.4.1.9.9.221.1.1.1.1.7",
		".1.3.6.1.4.1.9.9.221.1.1.1.1.18",
		".1.3.6.1.4.1.9.9.221.1.1.1.1.8",
		".1.3.6.1.4.1.9.9.221.1.1.1.1.20")
	if err == nil {
		return pools, err
	}

	return c.getMemoryComponentMemoryUsage(ctx,
		".1.3.6.1.4.1.9.9.48.1.1.1.2",
		".1.3.6.1.4.1.9.9.48.1.1.1.5",
		"",
		".1.3.6.1.4.1.9.9.48.1.1.1.6",
		"")
}

// GetMemoryComponentMemoryUsage returns the memory usage of ios devices.
func (c *iosCommunicator) getMemoryComponentMemoryUsage(ctx context.Context, poolLabelsOID, usedOID, usedHCOID, freeOID, freeHCOID network.OID) ([]device.MemoryPool, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	var pools []device.MemoryPool

	poolLabels, err := con.SNMP.SnmpClient.SNMPWalk(ctx, poolLabelsOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out memory pools")
	}

	for _, poolLabelVal := range poolLabels {
		poolLabel, err := poolLabelVal.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get memory pool label")
		}
		poolLabelString := poolLabel.String()

		idx := strings.Split(poolLabelVal.GetOID().String(), poolLabelsOID.String())[1]

		// get used value for memory pool
		used, err := c.getMemoryDecimalValue(ctx, con, usedOID.AddIndex(idx), usedHCOID.AddIndex(idx))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get used value for mempool '%s'", poolLabel)
		}

		// get free value for memory pool
		free, err := c.getMemoryDecimalValue(ctx, con, freeOID.AddIndex(idx), freeHCOID.AddIndex(idx))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get free value for mempool '%s'", poolLabel)
		}

		// usage = ( used / (free+used) ) * 100
		total := used.Add(free)
		if total.IsZero() {
			return nil, errors.New("total memory is zero, division by zero not possible")
		}

		usage, _ := used.DivRound(total, 4).Mul(decimal.NewFromInt(100)).Float64()

		var performanceDataPointModifier device.PerformanceDataPointModifier
		if strings.HasPrefix(poolLabelString, "lsmpi_io") {
			performanceDataPointModifier = func(p *monitoringplugin.PerformanceDataPoint) {
				p.SetThresholds(monitoringplugin.Thresholds{
					WarningMin:  nil,
					WarningMax:  nil,
					CriticalMin: 0,
					CriticalMax: 99.99,
				})
			}
		}

		pools = append(pools, device.MemoryPool{
			Label:                        &poolLabelString,
			Usage:                        &usage,
			PerformanceDataPointModifier: performanceDataPointModifier,
		})
	}

	return pools, nil
}

func (c *iosCommunicator) getMemoryDecimalValue(ctx context.Context, con *network.RequestDeviceConnection, oid, hcOid network.OID) (decimal.Decimal, error) {
	var snmpResponse network.SNMPResponse
	var hcRes []network.SNMPResponse
	var err error

	if hcOid != "" {
		hcRes, err = con.SNMP.SnmpClient.SNMPGet(ctx, hcOid)
	}
	if hcOid != "" && err == nil && len(hcRes) > 0 {
		snmpResponse = hcRes[0]
	} else {
		res, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
		if err != nil {
			return decimal.Decimal{}, errors.Wrap(err, "failed to read out hc and non hc oid")
		}
		if len(res) == 0 {
			return decimal.Decimal{}, errors.New("failed to read out hc and non hc value")
		}
		snmpResponse = res[0]
	}
	str, err := snmpResponse.GetValue()
	if err != nil {
		return decimal.Decimal{}, errors.Wrapf(err, "failed to convert value to string")
	}
	num, err := decimal.NewFromString(str.String())
	if err != nil {
		return decimal.Decimal{}, errors.Wrapf(err, "failed to convert value to decimal number")
	}

	return num, nil
}

func (c *iosCommunicator) GetHardwareHealthComponentPowerSupply(ctx context.Context) ([]device.HardwareHealthComponentPowerSupply, error) {
	res, err := c.deviceClass.GetHardwareHealthComponentPowerSupply(ctx)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= 3; i++ {
		chassisPS, err := c.getChassisPowerSupply(ctx, i)
		if err == nil {
			res = append(res, chassisPS)
		}
	}

	return res, nil
}

func (c *iosCommunicator) getChassisPowerSupply(ctx context.Context, id int) (device.HardwareHealthComponentPowerSupply, error) {
	var chassisPsXStatus network.OID
	switch id {
	case 1:
		chassisPsXStatus = "1.3.6.1.4.1.9.5.1.2.4"
	case 2:
		chassisPsXStatus = "1.3.6.1.4.1.9.5.1.2.7"
	case 3:
		chassisPsXStatus = "1.3.6.1.4.1.9.5.1.2.21"
	default:
		return device.HardwareHealthComponentPowerSupply{}, errors.New("invalid power supply id given")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return device.HardwareHealthComponentPowerSupply{}, errors.New("no device connection available")
	}

	response, err := con.SNMP.SnmpClient.SNMPWalk(ctx, chassisPsXStatus)
	if err != nil {
		return device.HardwareHealthComponentPowerSupply{}, errors.Wrap(err, "failed to get 'chassisPs3Status'")
	}

	if len(response) == 1 {
		val, err := response[0].GetValue()
		if err != nil {
			return device.HardwareHealthComponentPowerSupply{}, errors.Wrap(err, "failed to get 'chassisPs3Status' value")
		}
		var state device.HardwareHealthComponentState
		switch val.String() {
		// other
		case "1":
			// power supply not present
		// ok
		case "2":
			state = device.HardwareHealthComponentStateNormal
		// minorFault
		case "3":
			state = device.HardwareHealthComponentStateWarning
		// majorFault
		case "4":
			state = device.HardwareHealthComponentStateCritical
		}
		if state != "" {
			descr := "chassis_" + strconv.Itoa(id)
			return device.HardwareHealthComponentPowerSupply{
				Description: &descr,
				State:       &state,
			}, nil
		}
	}
	return device.HardwareHealthComponentPowerSupply{}, errors.New("power supply not found")
}
