package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type routerosCommunicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of routeros devices.
func (c *routerosCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}
	con.SNMP.SnmpClient.UseCache(false)

	interfaces, err := c.deviceClass.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	bitrateInResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.14988.1.1.1.2.1.9")
	if err != nil {
		log.Ctx(ctx).Warn().Msgf("snmpwalk for bitrate-in oid failed: %s", err)
	}

	bitratesIn := make(map[uint64]uint64)

	for _, bitrateInResult := range bitrateInResults {
		index, err := strconv.ParseUint(bitrateInResult.GetOID().GetIndex(), 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse oid port index to int (oid: %s)", bitrateInResult.GetOID())
		}
		if _, ok := bitratesIn[index]; ok {
			return nil, errors.New("multiple bitrate-in values found for index " + strconv.FormatUint(index, 10))
		}
		value, err := bitrateInResult.GetValue()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get bitrate-in value (oid: %s)", bitrateInResult.GetOID())
		}
		bitratesIn[index], err = value.UInt64()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get bitrate-in as int (oid: %s)", bitrateInResult.GetOID())
		}
	}

	bitrateOutResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.14988.1.1.1.2.1.8")
	if err != nil {
		log.Ctx(ctx).Warn().Msgf("snmpwalk for bitrate-out oid failed: %s", err)
	}

	bitratesOut := make(map[uint64]uint64)

	for _, bitrateOutResult := range bitrateOutResults {
		index, err := strconv.ParseUint(bitrateOutResult.GetOID().GetIndex(), 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse oid port index to int (oid: %s)", bitrateOutResult.GetOID())
		}
		if _, ok := bitratesOut[index]; ok {
			return nil, errors.New("multiple bitrate-out values found for index " + strconv.FormatUint(index, 10))
		}
		value, err := bitrateOutResult.GetValue()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get bitrate-out value (oid: %s)", bitrateOutResult.GetOID())
		}
		bitratesOut[index], err = value.UInt64()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get bitrate-out as int (oid: %s)", bitrateOutResult.GetOID())
		}
	}

	for i, iface := range interfaces {
		if iface.IfName != nil && strings.HasPrefix(*iface.IfName, "wlan") {
			bitrateIn, ok1 := bitratesIn[*iface.IfIndex]
			bitrateOut, ok2 := bitratesOut[*iface.IfIndex]

			if ok1 && ok2 {
				if iface.Radio != nil {
					interfaces[i].Radio.MaxbitrateIn = &bitrateIn
					interfaces[i].Radio.MaxbitrateIn = &bitrateOut
				} else {
					interfaces[i].Radio = &device.RadioInterface{
						MaxbitrateIn:  &bitrateIn,
						MaxbitrateOut: &bitrateOut,
					}
				}
				interfaces[i].MaxSpeedIn = &bitrateIn
				interfaces[i].MaxSpeedOut = &bitrateOut
			}
		}
	}

	return filterInterfaces(ctx, interfaces, filter)
}
