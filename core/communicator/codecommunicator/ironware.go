package codecommunicator

import (
	"context"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type ironwareCommunicator struct {
	codeCommunicator
}

// GetCPUComponentCPULoad returns the cpu load of ironware devices.
func (c *ironwareCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]float64, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}
	responses, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.1991.1.1.2.11.1.1.5")
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk failed")
	}
	var cpus []float64
	for _, response := range responses {
		if !strings.HasSuffix(response.GetOID(), "300") {
			continue
		}
		valueString, err := response.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		cpus = append(cpus, value)
	}
	return cpus, nil
}
