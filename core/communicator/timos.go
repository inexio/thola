package communicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
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

	descriptions, _ := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.6527.3.1.2.4.3.2.1.5")
	inbounds, _ := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.6527.6.2.2.2.8.1.1.1.4")
	outbounds, _ := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.6527.6.2.2.2.8.1.1.1.6")
	if len(inbounds) != len(descriptions) || len(outbounds) != len(descriptions) {
		return nil, errors.New("snmp tree lengths do not match")
	}
	for i, response := range descriptions {
		valueString, err := response.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		if valueString == "" {
			continue
		}
		in, err := inbounds[i].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		trafficIn, err := strconv.ParseUint(in, 0, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		out, err := outbounds[i].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		trafficOut, err := strconv.ParseUint(out, 0, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		interfaces = append(interfaces, device.Interface{
			IfDescr: &valueString,
			SAP:     &device.SAPInterface{Inbound: &trafficIn, Outbound: &trafficOut}})
	}

	return interfaces, nil
}
