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

	res, _ := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.9.9.109.1.1.1.1.5")
	for _, snmpres := range res {
		s, err := snmpres.GetValueString()
		if err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response to float64")
		}
		cpus = append(cpus, device.CPU{Load: &f})
	}
	return cpus, nil
}
