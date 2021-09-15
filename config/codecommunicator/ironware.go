package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type ironwareCommunicator struct {
	codeCommunicator
}

// GetCPUComponentCPULoad returns the cpu load of ironware devices.
func (c *ironwareCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]device.CPU, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	// snAgentCpuUtil100thPercent
	cpuUtilizationOID := ".1.3.6.1.4.1.1991.1.1.2.11.1.1.6"
	precision := 100.0
	cpuUtilization, err := con.SNMP.SnmpClient.SNMPWalk(ctx, cpuUtilizationOID)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("walking 'snAgentCpuUtil100thPercent' failed")

		// snAgentCpuUtilValue
		cpuUtilizationOID = ".1.3.6.1.4.1.1991.1.1.2.11.1.1.4"
		precision = 1.0
		cpuUtilization, err = con.SNMP.SnmpClient.SNMPWalk(ctx, cpuUtilizationOID)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("walking 'snAgentCpuUtilValue' failed")
			return nil, errors.Wrap(err, "getting CPU utilization failed")
		}
	}

	indexSlotNum := make(map[string]string)
	// snAgentCpuUtilSlotNum
	slotNumOID := ".1.3.6.1.4.1.1991.1.1.2.11.1.1.1"
	slotNum, err := con.SNMP.SnmpClient.SNMPWalk(ctx, slotNumOID)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("walking 'snAgentCpuUtilSlotNum' failed")
	} else {
		for _, num := range slotNum {
			res, err := num.GetValueString()
			if err != nil {
				return nil, errors.Wrap(err, "couldn't get string value")
			}
			indexSlotNum[strings.TrimPrefix(num.GetOID(), slotNumOID)] = res
		}
	}

	var cpus []device.CPU
	for _, cpuUtil := range cpuUtilization {
		if !strings.HasSuffix(cpuUtil.GetOID(), "300") {
			continue
		}
		valueString, err := cpuUtil.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		value /= precision
		cpu := device.CPU{Load: &value}
		if num, ok := indexSlotNum[strings.TrimPrefix(cpuUtil.GetOID(), cpuUtilizationOID)]; ok {
			cpu.Label = &num
		}
		cpus = append(cpus, cpu)
	}

	return cpus, nil
}
