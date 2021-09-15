package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"strconv"
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
		indices[cpuLoadResponse.GetOIDIndex()] = len(cpus) - 1 //current entry
	}

	// check deprecated cpu load oid. if one of the entries does not already exist in the cpu arr, add it
	for _, cpuLoadResponseDeprecated := range cpuLoad5minDeprecated {
		idx := cpuLoadResponseDeprecated.GetOIDIndex()

		if _, ok := indices[idx]; ok {
			continue
		}

		cpu, err := c.getCPUBySNMPResponse(cpuLoadResponseDeprecated)
		if err != nil {
			return nil, err
		}
		cpus = append(cpus, cpu)
		indices[cpuLoadResponseDeprecated.GetOIDIndex()] = len(cpus) - 1 //current entry
	}

	// read out physical indices for cpus
	physicalIndicesResult, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.2")
	if err != nil {
		// cannot determine cpu physical indices, return cpu loads without labels
		return cpus, nil
	}

	for _, physicalIndexResult := range physicalIndicesResult {
		idx := physicalIndexResult.GetOIDIndex()
		cpuIndex, ok := indices[idx]
		if !ok {
			continue
		}

		physicalIndex, err := physicalIndexResult.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get physical index as string")
		}

		// 0 == physical entry not supported
		if physicalIndex == "0" {
			continue
		}

		physicalNameResponse, err := con.SNMP.SnmpClient.SNMPGet(ctx, "1.3.6.1.2.1.47.1.1.1.1.7."+physicalIndex)
		if err != nil {
			// cannot get physical name, continue
			continue
		}

		physicalName, err := physicalNameResponse[0].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "physical name is not a string")
		}

		cpus[cpuIndex].Label = &physicalName
	}

	return cpus, nil
}

func (c *iosCommunicator) getCPUBySNMPResponse(res network.SNMPResponse) (device.CPU, error) {
	val, err := res.GetValueString()
	if err != nil {
		return device.CPU{}, errors.Wrap(err, "failed to get cpu load value")
	}
	valFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return device.CPU{}, errors.Wrap(err, "cpu load is not a float value")
	}
	return device.CPU{
		Label: nil,
		Load:  &valFloat,
	}, nil
}
