package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
)

type ekinopsModuleReaderAmplifier struct {
	ekinopsModuleData
}

func (m *ekinopsModuleReaderAmplifier) readModuleMetrics(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	var opticalAmplifierInterfaces []device.OpticalAmplifierInterface

	// booster ports
	oai, err := ekinopsReadAmplifierMetrics(ctx, ".1.3.6.1.4.1.20044.62.7.7.1.2", ".1.3.6.1.4.1.20044.62.9.4.1.1.3", ".1.3.6.1.4.1.20044.62.3.3.49.0", ".1.3.6.1.4.1.20044.62.3.3.50.0", ".1.3.6.1.4.1.20044.62.3.3.51.0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}
	opticalAmplifierInterfaces = append(opticalAmplifierInterfaces, oai...)

	// pre amp ports
	oai, err = ekinopsReadAmplifierMetrics(ctx, ".1.3.6.1.4.1.20044.62.7.8.1.2", ".1.3.6.1.4.1.20044.62.9.4.2.1.3", ".1.3.6.1.4.1.20044.62.3.2.33.0", ".1.3.6.1.4.1.20044.62.3.2.34.0", ".1.3.6.1.4.1.20044.62.3.2.35.0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}
	opticalAmplifierInterfaces = append(opticalAmplifierInterfaces, oai...)

	mappings, err := ekinopsInterfacesIfIdentifierToSliceIndex(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interface identifier mappings")
	}

	for _, opticalAmplifierInterface := range opticalAmplifierInterfaces {
		identifier := m.slotIdentifier + "/" + m.moduleName + "/" + *opticalAmplifierInterface.Identifier
		idx, ok := mappings[identifier]
		if !ok {
			return nil, fmt.Errorf("interface for identifier '%s' not found")
		}
		interfaces[idx].OpticalAmplifierInterface = opticalAmplifierInterface
	}
	return interfaces, nil
}

func ekinopsReadAmplifierMetrics(ctx context.Context, identifierOid, labelOid, txPowerOid, rxPowerOid, gainOid string) ([]device.OpticalAmplifierInterface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	identifierResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, identifierOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for identifier oid failed")
	}

	labelResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, labelOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for label oid failed")
	}

	var opticalAmplifierInterfaces []device.OpticalAmplifierInterface
	for k, identifierResult := range identifierResults {
		var opticalAmplifierInterface device.OpticalAmplifierInterface
		identifier, err := identifierResult.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get identifier for optical amplifier interface")
		}
		label, err := labelResults[k].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get label for optical amplifier interface")
		}
		opticalAmplifierInterface.Identifier = &identifier
		opticalAmplifierInterface.Label = &label
		opticalAmplifierInterfaces = append(opticalAmplifierInterfaces, opticalAmplifierInterface)
	}

	if c := len(opticalAmplifierInterfaces); c != 2 {
		return nil, fmt.Errorf("found %d optical amplifier interfaces in amplifier module, expected: 2", c)
	}

	// tx power
	txPowerResult, err := con.SNMP.SnmpClient.SNMPGet(ctx, txPowerOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for tx power oid failed")
	}
	avStr, err := txPowerResult[0].GetValueString()
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for tx power oid failed")
	}
	av, err := strconv.ParseFloat(avStr, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse snmp response to float64")
	}
	txPower := (av - 32768) * 0.005
	opticalAmplifierInterfaces[1].TXPower = &txPower

	// rx power
	rxPowerResult, err := con.SNMP.SnmpClient.SNMPGet(ctx, rxPowerOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for rx power oid failed")
	}
	avStr, err = rxPowerResult[0].GetValueString()
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for rx power oid failed")
	}
	av, err = strconv.ParseFloat(avStr, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse snmp response to float64")
	}
	rxPower := (av - 32768) * 0.005
	opticalAmplifierInterfaces[0].RXPower = &rxPower

	// gain
	gainResult, err := con.SNMP.SnmpClient.SNMPGet(ctx, gainOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for gain oid failed")
	}
	avStr, err = gainResult[0].GetValueString()
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for gain oid failed")
	}
	av, err = strconv.ParseFloat(avStr, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse gain to float64")
	}
	gain := (av - 32768) * 0.005
	opticalAmplifierInterfaces[1].Gain = &gain

	return opticalAmplifierInterfaces, nil
}
