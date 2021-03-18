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

	if advaGetDWDMInterfaces(ctx, interfaces) != nil {
		return nil, err
	}

	if advaGetChannelMatrix(ctx, interfaces) != nil {
		return nil, err
	}

	return interfaces, nil
}

func advaGetDWDMInterfaces(ctx context.Context, interfaces []device.Interface) error {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return errors.New("no device connection available")
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
		return errors.Wrap(err, "failed to read rx/tx power of ports")
	}

	for i, networkInterface := range interfaces {
		if specialValues, ok := specialInterfacesRaw[fmt.Sprint(*networkInterface.IfIndex)]; ok {
			err := addSpecialInterfacesValuesToInterface("dwdm", &interfaces[i], specialValues)
			if err != nil {
				log.Ctx(ctx).Trace().Err(err).Msg("can't parse oid values into Interface struct")
				return errors.Wrap(err, "can't parse oid values into Interface struct")
			}
		}
	}

	rx100Values, err := advaGetPowerValues(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.21")
	if err != nil {
		return errors.Wrap(err, "failed to get rx 100 values")
	}

	tx100Values, err := advaGetPowerValues(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.22")
	if err != nil {
		return errors.Wrap(err, "failed to get tx 100 values")
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
	}

	return nil
}

func advaGetChannelMatrix(ctx context.Context, interfaces []device.Interface) error {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return errors.New("no device connection available")
	}

	facilityPhysInstValueInputPower := ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.2"
	facilityPhysInstValueInputPowerValues, err := con.SNMP.SnmpClient.SNMPWalk(ctx, facilityPhysInstValueInputPower)
	if err != nil {
		return errors.Wrap(err, "failed to walk facilityPhysInstValueInputPower")
	}

	var subtrees []string
	var channels []device.OpticalChannel

	for _, res := range facilityPhysInstValueInputPowerValues {
		subtree := strings.TrimPrefix(res.GetOID(), facilityPhysInstValueInputPower)
		if s := strings.Split(strings.Trim(subtree, "."), "."); len(s) > 2 && s[len(s)-2] != "0" {
			val, err := res.GetValueString()
			if err != nil {
				return errors.Wrap(err, "failed to get rx value of channel "+subtree)
			}
			a, err := decimal.NewFromString(val)
			if err != nil {
				return errors.Wrap(err, "failed to parse rx value of channel "+subtree)
			}
			b := decimal.NewFromFloat(0.1)
			valFin, _ := a.Mul(b).Float64()

			subtrees = append(subtrees, subtree)
			channels = append(channels, device.OpticalChannel{
				Channel: s[len(s)-2],
				RXPower: &valFin,
			})
		}
	}

	for i, subtree := range subtrees {
		res, err := con.SNMP.SnmpClient.SNMPGet(ctx, ".1.3.6.1.4.1.2544.1.11.11.7.2.1.1.1.1"+subtree)
		if err != nil {
			return errors.Wrap(err, "failed to get facilityPhysInstValueOutputPower for subtree "+subtree)
		}

		if len(res) != 1 {
			return errors.New("failed to get tx value of channel " + subtree)
		}

		val, err := res[0].GetValueString()
		if err != nil {
			return errors.Wrap(err, "failed to get tx value of channel "+subtree)
		}
		a, err := decimal.NewFromString(val)
		if err != nil {
			return errors.Wrap(err, "failed to parse tx value of channel "+subtree)
		}
		b := decimal.NewFromFloat(0.1)
		valFin, _ := a.Mul(b).Float64()

		channels[i].TXPower = &valFin

		p := strings.Split(strings.ReplaceAll(strings.Trim(subtree, "."), "33152", "N"), ".")
		regex, err := regexp.Compile("-" + p[0] + "-" + p[1] + "-" + p[2])
		if err != nil {
			return errors.Wrap(err, "failed to build regex")
		}

		for j, interf := range interfaces {
			if interf.IfDescr != nil && regex.MatchString(*interf.IfDescr) {
				if interf.DWDM == nil {
					interfaces[j].DWDM = &device.DWDMInterface{}
				}
				interfaces[j].DWDM.Channels = append(interfaces[j].DWDM.Channels, channels[i])
				break
			}
		}
	}

	return nil
}

func advaGetPowerValues(ctx context.Context, oid string) (map[string]float64, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	values, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk facilityPhysInstValueCalculatedTotalPower")
	}

	descrToValues := make(map[string]float64)

	for _, val := range values {
		subtree := strings.TrimPrefix(val.GetOID(), oid)
		subtreeSplit := strings.Split(strings.Trim(subtree, "."), ".")
		if len(subtreeSplit) < 3 {
			return nil, errors.New("invalid facilityPhysInstValueCalculatedTotalPower value")
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
