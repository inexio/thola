package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	cpuUtilizationOID := network.OID(".1.3.6.1.4.1.1991.1.1.2.11.1.1.6")
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
	slotNumOID := network.OID(".1.3.6.1.4.1.1991.1.1.2.11.1.1.1")
	slotNum, err := con.SNMP.SnmpClient.SNMPWalk(ctx, slotNumOID)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("walking 'snAgentCpuUtilSlotNum' failed")
	} else {
		for _, num := range slotNum {
			res, err := num.GetValue()
			if err != nil {
				return nil, errors.Wrap(err, "couldn't get string value")
			}
			indexSlotNum[strings.TrimPrefix(num.GetOID().String(), slotNumOID.String())] = res.String()
		}
	}

	var cpus []device.CPU
	for _, cpuUtil := range cpuUtilization {
		if !strings.HasSuffix(cpuUtil.GetOID().String(), "300") {
			continue
		}
		value, err := cpuUtil.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		valueFloat, err := value.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		valueFloat /= precision
		cpu := device.CPU{Load: &valueFloat}
		if num, ok := indexSlotNum[strings.TrimPrefix(cpuUtil.GetOID().String(), cpuUtilizationOID.String())]; ok {
			cpu.Label = &num
		}
		cpus = append(cpus, cpu)
	}

	return cpus, nil
}
