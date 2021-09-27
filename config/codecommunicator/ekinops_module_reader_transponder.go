package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

type ekinopsModuleReaderTransponder struct {
	ekinopsModuleData
	networkLinePortsOIDs ekinopsTransponderOIDs
	clientPortsOIDs      ekinopsTransponderOIDs
}

type ekinopsTransponderOIDs struct {
	identifierOID  network.OID
	labelOID       network.OID
	txPowerOID     network.OID
	rxPowerOID     network.OID
	correctedFEC   network.OID
	uncorrectedFEC network.OID

	powerTransformFunc ekinopsPowerTransformFunc
}

func (m *ekinopsModuleReaderTransponder) readModuleMetrics(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	var OpticalTransponderInterfaces []device.OpticalTransponderInterface

	//  network / line ports
	oti, err := ekinopsReadTransponderMetrics(ctx, m.networkLinePortsOIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}
	OpticalTransponderInterfaces = append(OpticalTransponderInterfaces, oti...)

	// client ports
	oti, err = ekinopsReadTransponderMetrics(ctx, m.clientPortsOIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}
	OpticalTransponderInterfaces = append(OpticalTransponderInterfaces, oti...)

	mappings, err := ekinopsInterfacesIfIdentifierToSliceIndex(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interface identifier mappings")
	}

	for i, opticalTransponderInterface := range OpticalTransponderInterfaces {
		identifier := m.slotIdentifier + "/" + m.moduleName + "/" + strings.Split(*opticalTransponderInterface.Identifier, "(")[0]
		idx, ok := mappings[identifier]
		if !ok {
			log.Ctx(ctx).Debug().Msgf("interface for identifier '%s' not found", identifier)
			continue
		}
		interfaces[idx].OpticalTransponder = &OpticalTransponderInterfaces[i]
		interfaces[idx].IfAlias = interfaces[idx].OpticalTransponder.Label
	}
	return interfaces, nil
}

func ekinopsReadTransponderMetrics(ctx context.Context, oids ekinopsTransponderOIDs) ([]device.OpticalTransponderInterface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	if oids.identifierOID == "" || oids.labelOID == "" {
		return nil, errors.New("identifier and label oid need to be defined")
	}

	identifierResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oids.identifierOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for identifier oid failed")
	}

	labelResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oids.labelOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for label oid failed")
	}

	var txPowerResult, rxPowerResult, correctedFECResult, uncorrectedFECResult []network.SNMPResponse

	if oids.txPowerOID != "" {
		txPowerResult, err = con.SNMP.SnmpClient.SNMPWalk(ctx, oids.txPowerOID)
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk for tx power oid failed")
		}
	}

	if oids.rxPowerOID != "" {
		rxPowerResult, err = con.SNMP.SnmpClient.SNMPWalk(ctx, oids.rxPowerOID)
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk for rx power oid failed")
		}
	}

	if oids.correctedFEC != "" {
		correctedFECResult, err = con.SNMP.SnmpClient.SNMPWalk(ctx, oids.correctedFEC)
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk for corrected fec oid failed")
		}
	}

	if oids.uncorrectedFEC != "" {
		uncorrectedFECResult, err = con.SNMP.SnmpClient.SNMPWalk(ctx, oids.uncorrectedFEC)
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk for uncorrected fec oid failed")
		}
	}

	var opticalTransponderInterfaces []device.OpticalTransponderInterface
	for k, identifierResult := range identifierResults {
		var opticalTransponderInterface device.OpticalTransponderInterface
		identifier, err := identifierResult.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get identifier for optical amplifier interface")
		}
		identifierString := strings.Split(identifier.String(), "\n")[0]

		label, err := labelResults[k].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get label for optical amplifier interface")
		}
		labelString := strings.Split(label.String(), "\n")[0]

		opticalTransponderInterface.Identifier = &identifierString
		opticalTransponderInterface.Label = &labelString

		if rxPowerResult != nil {
			value, err := rxPowerResult[k].GetValue()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get rx power for optical amplifier interface")
			}
			valueFloat, err := value.Float64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse snmp response to float64")
			}

			if oids.powerTransformFunc != nil {
				valueFloat = oids.powerTransformFunc(valueFloat)
			}

			opticalTransponderInterface.RXPower = &valueFloat
		}

		if txPowerResult != nil {
			value, err := txPowerResult[k].GetValue()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get tx power for optical amplifier interface")
			}
			valueFloat, err := value.Float64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse snmp response to float64")
			}

			if oids.powerTransformFunc != nil {
				valueFloat = oids.powerTransformFunc(valueFloat)
			}

			opticalTransponderInterface.TXPower = &valueFloat
		}

		if correctedFECResult != nil {
			value, err := correctedFECResult[k].GetValue()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get corrected for optical amplifier interface")
			}
			valueUint, err := value.UInt64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse snmp response to uint64")
			}

			opticalTransponderInterface.CorrectedFEC = &valueUint
		}

		if uncorrectedFECResult != nil {
			value, err := uncorrectedFECResult[k].GetValue()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get uncorrected for optical amplifier interface")
			}
			valueUint, err := value.UInt64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse snmp response to uint64")
			}

			opticalTransponderInterface.UncorrectedFEC = &valueUint
		}
		opticalTransponderInterfaces = append(opticalTransponderInterfaces, opticalTransponderInterface)
	}

	return opticalTransponderInterfaces, nil
}
