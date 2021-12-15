package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type aviatCommunicator struct {
	codeCommunicator
}

func (c *aviatCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	var filterWithIfType []groupproperty.Filter
	for _, fil := range filter {
		if valueFilter, ok := fil.(groupproperty.ValueFilter); ok {
			if f := valueFilter.AddException([]string{"ifType"}); f != nil {
				filterWithIfType = append(filterWithIfType, f)
			}
		}
	}

	interfaces, err := c.deviceClass.GetInterfaces(ctx, filterWithIfType...)
	if err != nil {
		return nil, err
	}

	if groupproperty.CheckValueFiltersMatch(filter, []string{"radio"}) {
		log.Ctx(ctx).Debug().Msg("filter matched on 'radio', skipping aviat radio values")
		return filterInterfaces(ctx, interfaces, filter)
	}
	log.Ctx(ctx).Debug().Msg("reading aviat radio values")

	return c.getRadioInterface(ctx, interfaces, filter)
}

func (c *aviatCommunicator) getRadioInterface(ctx context.Context, interfaces []device.Interface, filter []groupproperty.Filter) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	var channels []device.RadioChannel

	// entPhysicalName
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.47.1.1.1.1.7")
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get entPhysicalName")
		return interfaces, nil
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
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatModemStatusMaxCapacity")
		return interfaces, nil
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
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatModemCurCapacityTx")
		return interfaces, nil
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
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatModemCurCapacityRx")
		return interfaces, nil
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
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatRxPerformRslReadingCurrent")
		return interfaces, nil
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
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatRxPerformTxpowReadingCurrent")
		return interfaces, nil
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

	// aviatRfFreqTx
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.5.2.1.1.1")
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatRfFreqTx")
		return interfaces, nil
	}

	for _, r := range res {
		txFrequencyVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatRfFreqTx value")
		}
		txFrequency, err := txFrequencyVal.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatRfFreqTx value")
		}
		txFrequency = txFrequency / 1000

		target := names[r.GetOID().GetIndex()]
		found := false
		for i, channel := range channels {
			if channel.Channel != nil && *channel.Channel == target {
				channels[i].TXFrequency = &txFrequency
				found = true
				break
			}
		}
		if !found {
			channels = append(channels, device.RadioChannel{
				Channel:     &target,
				TXFrequency: &txFrequency,
			})
		}
	}

	// aviatRfFreqRx
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2509.9.5.2.1.1.2")
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get aviatRfFreqRx")
		return interfaces, nil
	}

	for _, r := range res {
		rxFrequencyVal, err := r.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get aviatRfFreqRx value")
		}
		rxFrequency, err := rxFrequencyVal.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse aviatRfFreqRx value")
		}
		rxFrequency = rxFrequency / 1000

		target := names[r.GetOID().GetIndex()]
		found := false
		for i, channel := range channels {
			if channel.Channel != nil && *channel.Channel == target {
				channels[i].RXFrequency = &rxFrequency
				found = true
				break
			}
		}
		if !found {
			channels = append(channels, device.RadioChannel{
				Channel:     &target,
				RXFrequency: &rxFrequency,
			})
		}
	}

	for i, interf := range interfaces {
		if interf.IfType != nil && *interf.IfType == "radioMAC" {
			interfaces[i].MaxSpeedIn = &maxCapacity
			interfaces[i].MaxSpeedOut = &maxCapacity
			interfaces[i].Radio = &device.RadioInterface{
				MaxbitrateOut: &maxBitRateTx,
				MaxbitrateIn:  &maxBitRateRx,
				Channels:      channels,
			}
			break
		}
	}

	return filterInterfaces(ctx, interfaces, filter)
}
