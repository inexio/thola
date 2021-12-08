package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"strconv"
)

type aviatCommunicator struct {
	codeCommunicator
}

func (c *aviatCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	interfaces, err := c.deviceClass.GetInterfaces(ctx, filter...)
	if err != nil {
		return nil, err
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	var channels []device.RadioChannel

	// entPhysicalName
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.47.1.1.1.1.7")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aviatModemStatusMaxCapacity")
	}

	names := make(map[string]string)
	for _, r := range res {
		nameVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get entPhysicalName value")
		}
		names[r.GetOID().GetIndex()] = nameVal.String()
	}

	// aviatModemStatusMaxCapacity
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.3.2.4.1.1")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aviatModemStatusMaxCapacity")
	}

	var maxCapacity uint64
	for _, r := range res {
		capacityVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatModemStatusMaxCapacity value")
		}
		capacity, err := capacityVal.UInt64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatModemStatusMaxCapacity value")
		}
		maxCapacity += capacity
	}

	// aviatModemCurCapacityTx
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.3.2.1.1.11")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aviatModemCurCapacityTx")
	}

	var maxBitRateTx uint64
	for _, r := range res {
		bitRateVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatModemCurCapacityTx value")
		}
		bitRate, err := bitRateVal.UInt64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatModemCurCapacityTx value")
		}
		bitRate = bitRate * 1000
		maxBitRateTx += bitRate

		target := names[r.GetOID().GetIndex()]
		found := false
		for i, channel := range channels {
			if channel.Channel != nil && *channel.Channel == target {
				channels[i].MaxbitrateOut = &bitRate
				found = true
				break
			}
		}
		if !found {
			channels = append(channels, device.RadioChannel{
				Channel:       &target,
				MaxbitrateOut: &bitRate,
			})
		}
	}

	// aviatModemCurCapacityRx
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.3.2.1.1.12")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aviatModemCurCapacityRx")
	}

	var maxBitRateRx uint64
	for _, r := range res {
		bitRateVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatModemCurCapacityRx value")
		}
		bitRate, err := bitRateVal.UInt64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatModemCurCapacityRx value")
		}
		bitRate = bitRate * 1000
		maxBitRateRx += bitRate

		target := names[r.GetOID().GetIndex()]
		found := false
		for i, channel := range channels {
			if channel.Channel != nil && *channel.Channel == target {
				channels[i].MaxbitrateIn = &bitRate
				found = true
				break
			}
		}
		if !found {
			channels = append(channels, device.RadioChannel{
				Channel:      &target,
				MaxbitrateIn: &bitRate,
			})
		}
	}

	// aviatRxPerformRslReadingCurrent
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.15.2.2.1.4")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aviatRxPerformRslReadingCurrent")
	}

	for _, r := range res {
		levelInVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatRxPerformRslReadingCurrent value")
		}
		levelIn, err := levelInVal.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatRxPerformRslReadingCurrent value")
		}
		levelIn = levelIn / 10

		target := names[r.GetOID().GetIndex()]
		found := false
		for i, channel := range channels {
			if channel.Channel != nil && *channel.Channel == target {
				channels[i].LevelIn = &levelIn
				found = true
				break
			}
		}
		if !found {
			channels = append(channels, device.RadioChannel{
				Channel: &target,
				LevelIn: &levelIn,
			})
		}
	}

	// aviatRxPerformTxpowReadingCurrent
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.33.2.2.1.7")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aviatRxPerformTxpowReadingCurrent")
	}

	for _, r := range res {
		levelOutVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatRxPerformTxpowReadingCurrent value")
		}
		levelOut, err := levelOutVal.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatRxPerformTxpowReadingCurrent value")
		}
		levelOut = levelOut / 10

		target := names[r.GetOID().GetIndex()]
		found := false
		for i, channel := range channels {
			if channel.Channel != nil && *channel.Channel == target {
				channels[i].LevelOut = &levelOut
				found = true
				break
			}
		}
		if !found {
			channels = append(channels, device.RadioChannel{
				Channel:  &target,
				LevelOut: &levelOut,
			})
		}
	}

	var radioIfIndex *uint64

	// ifType
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.2.2.1.3")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ifType")
	}

	for _, r := range res {
		ifTypeVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get ifType value")
		}

		if ifTypeVal.String() == "188" {
			ifIndex, err := strconv.ParseUint(r.GetOID().GetIndex(), 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse ifIndex value")
			}
			radioIfIndex = &ifIndex
		}
	}

	for i, interf := range interfaces {
		if interf.IfIndex != nil && radioIfIndex != nil && *interf.IfIndex == *radioIfIndex {
			interfaces[i].MaxSpeedIn = &maxCapacity
			interfaces[i].MaxSpeedOut = &maxCapacity
			interfaces[i].Radio = &device.RadioInterface{
				MaxbitrateOut: &maxBitRateTx,
				MaxbitrateIn:  &maxBitRateRx,
				Channels:      channels,
			}
		}
	}

	return interfaces, nil
}
