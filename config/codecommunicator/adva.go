package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

type advaCommunicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of adva devices.
func (c *advaCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	interfaces, err := c.parent.GetInterfaces(ctx, filter...)
	if err != nil {
		return nil, err
	}

	if groupproperty.CheckValueFiltersMatch(filter, []string{"dwdm"}) {
		log.Ctx(ctx).Debug().Msg("filter matched on 'dwdm', skipping adva dwdm values")
		return c.normalizeInterfaces(ctx, interfaces, filter)
	}
	log.Ctx(ctx).Debug().Msg("reading adva dwdm values")

	if err = c.getDWDMInterfaces(ctx, interfaces); err != nil {
		return nil, err
	}

	if err = c.getChannels(ctx, interfaces); err != nil {
		return nil, err
	}

	return c.normalizeInterfaces(ctx, interfaces, filter)
}

func (c *advaCommunicator) getDWDMInterfaces(ctx context.Context, interfaces []device.Interface) error {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return errors.New("no device connection available")
	}

	rxPowerRaw, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2544.1.11.2.4.3.5.1.3")
	if err != nil {
		return errors.Wrap(err, "failed to walk rx power")
	}

	rxPower := make(map[string]float64)

	for _, resp := range rxPowerRaw {
		res, err := resp.GetValue()
		if err != nil {
			return errors.Wrap(err, "failed to convert rx power to string")
		}
		rxValue, err := decimal.NewFromString(res.String())
		if err != nil {
			return errors.Wrap(err, "failed to convert rx power to decimal")
		}
		oid := strings.Split(resp.GetOID().String(), ".")
		rxPower[oid[len(oid)-1]], _ = rxValue.Mul(decimal.NewFromFloat(0.1)).Float64()
	}

	txPowerRaw, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2544.1.11.2.4.3.5.1.4")
	if err != nil {
		return errors.Wrap(err, "failed to walk tx power")
	}

	txPower := make(map[string]float64)

	for _, resp := range txPowerRaw {
		res, err := resp.GetValue()
		if err != nil {
			return errors.Wrap(err, "failed to convert tx power to string")
		}
		txValue, err := decimal.NewFromString(res.String())
		if err != nil {
			return errors.Wrap(err, "failed to convert tx power to decimal")
		}
		oid := strings.Split(resp.GetOID().String(), ".")
		txPower[oid[len(oid)-1]], _ = txValue.Mul(decimal.NewFromFloat(0.1)).Float64()
	}

	rx100Values, err := c.getPowerValues(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.21")
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get rx 100 values")
	}

	tx100Values, err := c.getPowerValues(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.22")
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get tx 100 values")
	}

	for i, interf := range interfaces {
		if interf.IfIndex != nil {
			// rx power
			if value, ok := rxPower[fmt.Sprint(*interf.IfIndex)]; ok {
				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}
				interfaces[i].DWDM.RXPower = &value
			}

			// tx power
			if value, ok := txPower[fmt.Sprint(*interf.IfIndex)]; ok {
				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}
				interfaces[i].DWDM.TXPower = &value
			}

			// corrected fec 15m
			res, err := con.SNMP.SnmpClient.SNMPGet(ctx, network.OID(".1.3.6.1.4.1.2544.1.11.2.6.2.180.1.2.").AddIndex(fmt.Sprint(*interf.IfIndex)+".1"))
			if err == nil && len(res) == 1 {
				val, err := res[0].GetValue()
				if err != nil {
					return errors.Wrap(err, "failed to get corrected 15m bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				valFloat, err := val.Float64()
				if err != nil {
					return errors.Wrap(err, "failed to parse corrected 15m bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.CorrectedFEC = append(interfaces[i].DWDM.CorrectedFEC, device.Rate{
					Time:  "15m",
					Value: valFloat,
				})
			}

			// uncorrected fec 15m
			res, err = con.SNMP.SnmpClient.SNMPGet(ctx, network.OID(".1.3.6.1.4.1.2544.1.11.2.6.2.180.1.3.").AddIndex(fmt.Sprint(*interf.IfIndex)+".1"))
			if err == nil && len(res) == 1 {
				val, err := res[0].GetValue()
				if err != nil {
					return errors.Wrap(err, "failed to get uncorrected 15m bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				valFloat, err := val.Float64()
				if err != nil {
					return errors.Wrap(err, "failed to parse uncorrected 15m bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.UncorrectedFEC = append(interfaces[i].DWDM.UncorrectedFEC, device.Rate{
					Time:  "15m",
					Value: valFloat,
				})
			}

			// corrected fec 1d
			res, err = con.SNMP.SnmpClient.SNMPGet(ctx, network.OID(".1.3.6.1.4.1.2544.1.11.2.6.2.181.1.2.").AddIndex(fmt.Sprint(*interf.IfIndex)+".1"))
			if err == nil && len(res) == 1 {
				val, err := res[0].GetValue()
				if err != nil {
					return errors.Wrap(err, "failed to get corrected 1d bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				valFloat, err := val.Float64()
				if err != nil {
					return errors.Wrap(err, "failed to parse corrected 1d bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.CorrectedFEC = append(interfaces[i].DWDM.CorrectedFEC, device.Rate{
					Time:  "1d",
					Value: valFloat,
				})
			}

			// uncorrected fec 1d
			res, err = con.SNMP.SnmpClient.SNMPGet(ctx, network.OID(".1.3.6.1.4.1.2544.1.11.2.6.2.181.1.3.").AddIndex(fmt.Sprint(*interf.IfIndex)+".1"))
			if err == nil && len(res) == 1 {
				valFloat, err := res[0].GetValue()
				if err != nil {
					return errors.Wrap(err, "failed to get uncorrected 1d bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				val, err := valFloat.Float64()
				if err != nil {
					return errors.Wrap(err, "failed to parse uncorrected 1d bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.UncorrectedFEC = append(interfaces[i].DWDM.UncorrectedFEC, device.Rate{
					Time:  "1d",
					Value: val,
				})
			}
		}

		// overwrite rx/tx power for 100g interfaces
		if interf.IfDescr != nil {
			rxVal, rxOK := rx100Values[*interf.IfDescr]
			txVal, txOK := tx100Values[*interf.IfDescr]
			if (rxOK || txOK) && interf.DWDM == nil {
				interfaces[i].DWDM = &device.DWDMInterface{}
			}
			if rxOK && (interfaces[i].DWDM.RXPower == nil || *interfaces[i].DWDM.RXPower == -6553.5) {
				interfaces[i].DWDM.RXPower = &rxVal
			}
			if txOK && (interfaces[i].DWDM.TXPower == nil || *interfaces[i].DWDM.TXPower == -6553.5) {
				interfaces[i].DWDM.TXPower = &txVal
			}
		}
	}

	return nil
}

func (c *advaCommunicator) getChannels(ctx context.Context, interfaces []device.Interface) error {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return errors.New("no device connection available")
	}

	channels := make(map[string]device.OpticalChannel)

	facilityPhysInstValueInputPower := network.OID(".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.2")
	facilityPhysInstValueInputPowerValues, err := con.SNMP.SnmpClient.SNMPWalk(ctx, facilityPhysInstValueInputPower)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to walk facilityPhysInstValueInputPower")
	}

	for _, res := range facilityPhysInstValueInputPowerValues {
		subtree := strings.TrimPrefix(res.GetOID().String(), facilityPhysInstValueInputPower.String())
		if s := strings.Split(strings.Trim(subtree, "."), "."); len(s) > 3 && s[len(s)-2] != "0" && s[len(s)-3] == "33152" {
			val, err := res.GetValue()
			if err != nil {
				return errors.Wrap(err, "failed to get rx value of channel "+subtree)
			}
			a, err := decimal.NewFromString(val.String())
			if err != nil {
				return errors.Wrap(err, "failed to parse rx value of channel "+subtree)
			}
			b := decimal.NewFromFloat(0.1)
			valFin, _ := a.Mul(b).Float64()

			channels[subtree] = device.OpticalChannel{
				Channel: &s[len(s)-2],
				RXPower: &valFin,
			}
		}
	}

	facilityPhysInstValueOutputPower := network.OID(".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.1")
	facilityPhysInstValueOutputPowerValues, err := con.SNMP.SnmpClient.SNMPWalk(ctx, facilityPhysInstValueOutputPower)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to walk facilityPhysInstValueOutputPower")
	}

	for _, res := range facilityPhysInstValueOutputPowerValues {
		subtree := strings.TrimPrefix(res.GetOID().String(), facilityPhysInstValueOutputPower.String())
		if s := strings.Split(strings.Trim(subtree, "."), "."); len(s) > 3 && s[len(s)-2] != "0" && s[len(s)-3] == "33152" {
			val, err := res.GetValue()
			if err != nil {
				return errors.Wrap(err, "failed to get tx value of channel "+subtree)
			}
			a, err := decimal.NewFromString(val.String())
			if err != nil {
				return errors.Wrap(err, "failed to parse tx value of channel "+subtree)
			}
			b := decimal.NewFromFloat(0.1)
			valFin, _ := a.Mul(b).Float64()

			if channel, ok := channels[subtree]; !ok {
				channels[subtree] = device.OpticalChannel{
					Channel: &s[len(s)-2],
					TXPower: &valFin,
				}
			} else {
				channel.TXPower = &valFin
				channels[subtree] = channel
			}
		}
	}

	subtype := "channelMonitoring"

	for subtree, channel := range channels {
		s := strings.Split(strings.Trim(subtree, "."), ".")
		for j, interf := range interfaces {
			if interf.IfDescr != nil && strings.Contains(*interf.IfDescr, "-"+s[0]+"-"+s[1]+"-N") {
				if interf.DWDM == nil {
					interfaces[j].DWDM = &device.DWDMInterface{}
				}
				interfaces[j].DWDM.Channels = append(interfaces[j].DWDM.Channels, channel)
				interfaces[j].SubType = &subtype
				break
			}
		}
	}

	return nil
}

func (c *advaCommunicator) getPowerValues(ctx context.Context, oid network.OID) (map[string]float64, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	values, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk "+oid.String())
	}

	descrToValues := make(map[string]float64)

	for _, val := range values {
		subtree := strings.TrimPrefix(val.GetOID().String(), oid.String())
		subtreeSplit := strings.Split(strings.Trim(subtree, "."), ".")
		if len(subtreeSplit) < 3 {
			return nil, errors.New("invalid value for oid " + oid.String())
		}

		value, err := val.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get rx value")
		}
		valueDecimal, err := decimal.NewFromString(value.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse rx value")
		}
		multiplier := decimal.NewFromFloat(0.1)

		portInt, err := strconv.Atoi(subtreeSplit[2])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse oid to int")
		}

		descrToValues["CH-"+subtreeSplit[0]+"-"+subtreeSplit[1]+"-C"+strconv.Itoa(portInt/256)], _ = valueDecimal.Mul(multiplier).Float64()
	}

	return descrToValues, nil
}

func (c *advaCommunicator) normalizeInterfaces(ctx context.Context, interfaces []device.Interface, filter []groupproperty.Filter) ([]device.Interface, error) {
	return filterInterfaces(ctx, interfaces, append(filter, groupproperty.GetGroupFilter([]string{"ifDescr"}, "^(TIFI-|TIFO-)")))
}
