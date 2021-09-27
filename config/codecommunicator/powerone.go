package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
)

type poweroneACCCommunicator struct {
	codeCommunicator
}

type poweronePCCCommunicator struct {
	codeCommunicator
}

// GetUPSComponentMainsVoltageApplied returns the ups state of powerone/acc devices.
func (c *poweroneACCCommunicator) GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error) {
	return getPoweroneMainsVoltageApplied(ctx, ".1.3.6.1.4.1.5961.4.3.2.0")
}

// GetUPSComponentMainsVoltageApplied returns the ups state of powerone/pcc devices.
func (c *poweronePCCCommunicator) GetUPSComponentMainsVoltageApplied(ctx context.Context) (bool, error) {
	return getPoweroneMainsVoltageApplied(ctx, ".1.3.6.1.4.1.5961.3.3.2.0")
}

func getPoweroneMainsVoltageApplied(ctx context.Context, oid network.OID) (bool, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return false, errors.New("no device connection available")
	}
	response, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
	if err != nil {
		return false, errors.Wrap(err, "snmpget failed")
	}
	if len(response) != 1 {
		return false, errors.New("no or more than one snmp response available")
	}
	val, err := response[0].GetValue()
	if err != nil {
		return false, errors.Wrap(err, "couldn't get string value")
	}
	value, err := val.Int()
	if err != nil {
		return false, errors.Wrap(err, "failed to parse snmp response")
	}
	return (value & 8) == 0, nil
}
