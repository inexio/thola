package communicator

import (
	"context"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
)

type iosCommunicator struct {
	baseCommunicator
}

func (c *iosCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]float64, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}
	var cpus []float64

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
		cpus = append(cpus, f)
	}
	return cpus, nil
}
