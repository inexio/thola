package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

type ekinopsModuleReaderAmplifier struct {
	ekinopsModuleData
	boosterPorts ekinopsAmplifierOIDs
	preAmpPorts  ekinopsAmplifierOIDs
}

type ekinopsAmplifierOIDs struct {
	identifierOID network.OID
	labelOID      network.OID
	txPowerOID    network.OID
	rxPowerOID    network.OID
	gainOID       network.OID

	powerTransformFunc ekinopsPowerTransformFunc
}

func (m *ekinopsModuleReaderAmplifier) readModuleMetrics(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	var opticalAmplifierInterfaces []device.OpticalAmplifierInterface

	// booster ports
	oai, err := ekinopsReadAmplifierMetrics(ctx, m.boosterPorts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}
	opticalAmplifierInterfaces = append(opticalAmplifierInterfaces, oai...)

	// pre amp ports
	oai, err = ekinopsReadAmplifierMetrics(ctx, m.preAmpPorts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for pre amp ports")
	}
	opticalAmplifierInterfaces = append(opticalAmplifierInterfaces, oai...)

	mappings, err := ekinopsInterfacesIfIdentifierToSliceIndex(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interface identifier mappings")
	}

	for i, opticalAmplifierInterface := range opticalAmplifierInterfaces {
		identifier := m.slotIdentifier + "/" + m.moduleName + "/" + strings.Split(*opticalAmplifierInterface.Identifier, "(")[0]
		idx, ok := mappings[identifier]
		if !ok {
			log.Ctx(ctx).Debug().Msgf("interface for identifier '%s' not found", identifier)
			continue
		}
		interfaces[idx].OpticalAmplifier = &opticalAmplifierInterfaces[i]
		interfaces[idx].IfAlias = interfaces[idx].OpticalAmplifier.Label
	}
	return interfaces, nil
}

func ekinopsReadAmplifierMetrics(ctx context.Context, oids ekinopsAmplifierOIDs) ([]device.OpticalAmplifierInterface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	identifierResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oids.identifierOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for identifier oid failed")
	}

	labelResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oids.labelOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for label oid failed")
	}

	var opticalAmplifierInterfaces []device.OpticalAmplifierInterface
	for k, identifierResult := range identifierResults {
		var opticalAmplifierInterface device.OpticalAmplifierInterface
		identifier, err := identifierResult.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get identifier for optical amplifier interface")
		}
		identifierString := identifier.String()
		label, err := labelResults[k].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get label for optical amplifier interface")
		}
		labelString := label.String()
		opticalAmplifierInterface.Identifier = &identifierString
		opticalAmplifierInterface.Label = &labelString
		opticalAmplifierInterfaces = append(opticalAmplifierInterfaces, opticalAmplifierInterface)
	}

	if c := len(opticalAmplifierInterfaces); c != 2 {
		return nil, fmt.Errorf("found %d optical amplifier interfaces in amplifier module, expected: 2", c)
	}

	// tx power
	txPowerResult, err := con.SNMP.SnmpClient.SNMPGet(ctx, oids.txPowerOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for tx power oid failed")
	}
	avVal, err := txPowerResult[0].GetValue()
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for tx power oid failed")
	}
	av, err := avVal.Float64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse snmp response to float64")
	}
	txPower := oids.powerTransformFunc(av)
	opticalAmplifierInterfaces[1].TXPower = &txPower

	// rx power
	rxPowerResult, err := con.SNMP.SnmpClient.SNMPGet(ctx, oids.rxPowerOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for rx power oid failed")
	}
	avVal, err = rxPowerResult[0].GetValue()
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for rx power oid failed")
	}
	av, err = avVal.Float64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse snmp response to float64")
	}
	rxPower := oids.powerTransformFunc(av)
	opticalAmplifierInterfaces[0].RXPower = &rxPower

	// gain
	gainResult, err := con.SNMP.SnmpClient.SNMPGet(ctx, oids.gainOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for gain oid failed")
	}
	avVal, err = gainResult[0].GetValue()
	if err != nil {
		return nil, errors.Wrap(err, "snmpget failed for gain oid failed")
	}
	av, err = avVal.Float64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse gain to float64")
	}
	gain := oids.powerTransformFunc(av)
	opticalAmplifierInterfaces[1].Gain = &gain

	return opticalAmplifierInterfaces, nil
}
