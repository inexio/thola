package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/inexio/thola/core/value"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"regexp"
	"strconv"
	"strings"
)

type advaCommunicator struct {
	baseCommunicator
}

// GetInterfaces returns the interfaces of adva devices.
func (c *advaCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	interfaces, err := c.sub.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	if interfaces, err = advaGetDWDMInterfaces(ctx, interfaces); err != nil {
		return nil, err
	}

	if interfaces, err = advaGetChannels(ctx, interfaces); err != nil {
		return nil, err
	}

	if interfaces, err = advaGet100GInterfaces(ctx, interfaces); err != nil {
		return nil, err
	}

	return interfaces, nil
}

func advaGetDWDMInterfaces(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	specialInterfacesRaw, err := getValuesBySNMPWalk(ctx, deviceClassOIDs{
		"rx_power": deviceClassOID{
			SNMPGetConfiguration: network.SNMPGetConfiguration{
				OID: "1.3.6.1.4.1.2544.1.11.2.4.3.5.1.3",
			},
			operators: propertyOperators{
				&modifyOperatorAdapter{
					&multiplyNumberModifier{
						value: &constantPropertyReader{
							Value: value.New("0.1"),
						},
					},
				},
			},
		},
		"tx_power": deviceClassOID{
			SNMPGetConfiguration: network.SNMPGetConfiguration{
				OID: "1.3.6.1.4.1.2544.1.11.2.4.3.5.1.4",
			},
			operators: propertyOperators{
				&modifyOperatorAdapter{
					&multiplyNumberModifier{
						value: &constantPropertyReader{
							Value: value.New("0.1"),
						},
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to read rx/tx power of ports")
	}

	for i, networkInterface := range interfaces {
		if specialValues, ok := specialInterfacesRaw[fmt.Sprint(*networkInterface.IfIndex)]; ok {
			err := addSpecialInterfacesValuesToInterface("dwdm", &interfaces[i], specialValues)
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("can't parse oid values into Interface struct")
				return nil, errors.Wrap(err, "can't parse oid values into Interface struct")
			}
		}
	}

	rx100Values, err := advaGetPowerValues(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.21")
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get rx 100 values")
	}

	tx100Values, err := advaGetPowerValues(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.22")
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get tx 100 values")
	}

	for i, interf := range interfaces {
		if interf.IfDescr != nil {
			rxVal, rxOK := rx100Values[*interf.IfDescr]
			txVal, txOK := tx100Values[*interf.IfDescr]
			if (rxOK || txOK) && interf.DWDM == nil {
				interfaces[i].DWDM = &device.DWDMInterface{}
			}
			if rxOK {
				interfaces[i].DWDM.RXPower100G = &rxVal
			}
			if txOK {
				interfaces[i].DWDM.TXPower100G = &txVal
			}
		}

		if interf.IfIndex != nil {
			// corrected fec 15m
			res, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.2.6.2.180.1.2."+fmt.Sprint(*interf.IfIndex)+".1")
			if err == nil && len(res) == 1 {
				valString, err := res[0].GetValueString()
				if err != nil {
					return nil, errors.Wrap(err, "failed to get corrected 15m bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				val, err := strconv.ParseFloat(valString, 64)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse corrected 15m bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.CorrectedFEC = append(interfaces[i].DWDM.CorrectedFEC, device.Rate{
					Time:  "15m",
					Value: val,
				})
			}

			// uncorrected fec 15m
			res, err = con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.2.6.2.180.1.3."+fmt.Sprint(*interf.IfIndex)+".1")
			if err == nil && len(res) == 1 {
				valString, err := res[0].GetValueString()
				if err != nil {
					return nil, errors.Wrap(err, "failed to get uncorrected 15m bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				val, err := strconv.ParseFloat(valString, 64)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse uncorrected 15m bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.UncorrectedFEC = append(interfaces[i].DWDM.UncorrectedFEC, device.Rate{
					Time:  "15m",
					Value: val,
				})
			}

			// corrected fec 1d
			res, err = con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.2.6.2.181.1.2."+fmt.Sprint(*interf.IfIndex)+".1")
			if err == nil && len(res) == 1 {
				valString, err := res[0].GetValueString()
				if err != nil {
					return nil, errors.Wrap(err, "failed to get corrected 1d bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				val, err := strconv.ParseFloat(valString, 64)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse corrected 1d bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
				}

				if interfaces[i].DWDM == nil {
					interfaces[i].DWDM = &device.DWDMInterface{}
				}

				interfaces[i].DWDM.CorrectedFEC = append(interfaces[i].DWDM.CorrectedFEC, device.Rate{
					Time:  "1d",
					Value: val,
				})
			}

			// uncorrected fec 1d
			res, err = con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.2.6.2.181.1.3."+fmt.Sprint(*interf.IfIndex)+".1")
			if err == nil && len(res) == 1 {
				valString, err := res[0].GetValueString()
				if err != nil {
					return nil, errors.Wrap(err, "failed to get uncorrected 1d bit error rate string value for interface "+fmt.Sprint(*interf.IfIndex))
				}

				val, err := strconv.ParseFloat(valString, 64)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse uncorrected 1d bit error rate for interface "+fmt.Sprint(*interf.IfIndex))
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
	}

	return interfaces, nil
}

func advaGetChannels(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	facilityPhysInstValueInputPower := ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.2"
	facilityPhysInstValueInputPowerValues, err := con.SNMP.SnmpClient.SNMPWalk(ctx, facilityPhysInstValueInputPower)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to walk facilityPhysInstValueInputPower")
	}

	var subtrees []string
	channels := make(map[string]device.OpticalChannel)
	subtype := "channelMonitoring"

	for _, res := range facilityPhysInstValueInputPowerValues {
		subtree := strings.TrimPrefix(res.GetOID(), facilityPhysInstValueInputPower)
		if s := strings.Split(strings.Trim(subtree, "."), "."); len(s) > 2 && s[len(s)-2] != "0" {
			val, err := res.GetValueString()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get rx value of channel "+subtree)
			}
			a, err := decimal.NewFromString(val)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse rx value of channel "+subtree)
			}
			b := decimal.NewFromFloat(0.1)
			valFin, _ := a.Mul(b).Float64()

			subtrees = append(subtrees, subtree)
			channels[subtree] = device.OpticalChannel{
				Channel: s[len(s)-2],
				RXPower: &valFin,
			}
		}
	}

	for _, subtree := range subtrees {
		res, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.1"+subtree)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get facilityPhysInstValueOutputPower for subtree "+subtree)
		}

		if len(res) != 1 {
			return nil, errors.New("failed to get tx value of subtree " + subtree)
		}

		val, err := res[0].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get tx value of subtree "+subtree)
		}
		valueDecimal, err := decimal.NewFromString(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse tx value of subtree "+subtree)
		}
		multiplier := decimal.NewFromFloat(0.1)
		valFin, _ := valueDecimal.Mul(multiplier).Float64()

		channel := channels[subtree]
		channel.TXPower = &valFin

		p := strings.Split(strings.ReplaceAll(strings.Trim(subtree, "."), "33152", "N"), ".")
		if len(p) < 3 {
			return nil, errors.New("invalid channel identifier")
		}
		regex, err := regexp.Compile("-" + p[0] + "-" + p[1] + "-" + p[2])
		if err != nil {
			return nil, errors.Wrap(err, "failed to build regex")
		}

		for j, interf := range interfaces {
			if interf.IfDescr != nil && regex.MatchString(*interf.IfDescr) {
				if interf.DWDM == nil {
					interfaces[j].DWDM = &device.DWDMInterface{}
				}
				interfaces[j].DWDM.Channels = append(interfaces[j].DWDM.Channels, channel)
				interfaces[j].SubType = &subtype
				break
			}
		}
	}

	return interfaces, nil
}

func advaGet100GInterfaces(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	ports := ".1.3.6.1.4.1.2544.1.11.7.2.7.1.6"
	portValues, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ports)
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to walk 100g ports")
	}

	var subtrees []string
	otlInterfaces := make(map[string]device.Interface)
	subType := "OTL"

	for _, res := range portValues {
		portName, err := res.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get snmp response")
		}

		if strings.HasPrefix(portName, "OTL-") {
			subtree := strings.TrimPrefix(res.GetOID(), ports)

			rxValue, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.7.7.2.3.1.2"+subtree)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get rx value of port "+portName)
			}

			if len(rxValue) != 1 {
				return nil, errors.Wrap(err, "failed to get rx value of port "+portName)
			}

			rxValueString, err := rxValue[0].GetValueString()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get snmp response")
			}

			valueDecimal, err := decimal.NewFromString(rxValueString)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse rx value of subtree "+subtree)
			}
			multiplier := decimal.NewFromFloat(0.1)
			valFin, _ := valueDecimal.Mul(multiplier).Float64()

			subtrees = append(subtrees, subtree)
			otlInterfaces[subtree] = device.Interface{
				IfDescr: &portName,
				SubType: &subType,
				DWDM: &device.DWDMInterface{
					RXPower: &valFin,
				},
			}
		}
	}

	for _, subtree := range subtrees {
		res, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.7.7.2.3.1.1"+subtree)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get tx value for subtree "+subtree)
		}

		if len(res) != 1 {
			return nil, errors.New("failed to get tx value of subtree " + subtree)
		}

		val, err := res[0].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get tx value of subtree "+subtree)
		}
		a, err := decimal.NewFromString(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse tx value of subtree "+subtree)
		}
		b := decimal.NewFromFloat(0.1)
		valFin, _ := a.Mul(b).Float64()

		interf := otlInterfaces[subtree]
		interf.DWDM.TXPower = &valFin

		interfaces = append(interfaces, interf)
	}

	return interfaces, nil
}

func advaGetPowerValues(ctx context.Context, oid string) (map[string]float64, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	values, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk "+oid)
	}

	descrToValues := make(map[string]float64)

	for _, val := range values {
		subtree := strings.TrimPrefix(val.GetOID(), oid)
		subtreeSplit := strings.Split(strings.Trim(subtree, "."), ".")
		if len(subtreeSplit) < 3 {
			return nil, errors.New("invalid value for oid " + oid)
		}

		valueString, err := val.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get rx value")
		}
		valueDecimal, err := decimal.NewFromString(valueString)
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
