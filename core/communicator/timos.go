package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type timosCommunicator struct {
	baseCommunicator
}

// GetInterfaces returns the interfaces of timetra devices.
func (c *timosCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	interfaces, err := c.sub.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	ports := ".1.3.6.1.4.1.6527.3.1.2.4.3.2.1.5"
	descriptions, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ports)
	if err != nil {
		return nil, errors.Wrap(err, "error during snmp walk")
	}
	for _, response := range descriptions {
		valueString, err := response.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		if valueString == "" {
			continue
		}
		index := strings.TrimPrefix(response.GetOID(), ports)
		in, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.6527.6.2.2.2.8.1.1.1.4"+index)
		if err != nil {
			return nil, errors.Wrap(err, "error during snmp get")
		}
		inStr, err := in[0].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		trafficIn, err := strconv.ParseUint(inStr, 0, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		out, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.6527.6.2.2.2.8.1.1.1.6"+index)
		if err != nil {
			return nil, errors.Wrap(err, "error during snmp get")
		}
		outStr, err := out[0].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		trafficOut, err := strconv.ParseUint(outStr, 0, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		interfaces = append(interfaces, device.Interface{
			IfDescr: &valueString,
			SAP:     &device.SAPInterface{Inbound: &trafficIn, Outbound: &trafficOut}})
	}

	return interfaces, nil
}
