package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type ekinopsModuleReaderOPM8 struct {
	ekinopsModuleData
	ports ekinopsOPMOIDs
}

type ekinopsOPMOIDs struct {
	identifierOID string
	labelOID      string
	rxPowerOID    string
	channelsOid   string

	powerTransformFunc ekinopsPowerTransformFunc
}

func (m *ekinopsModuleReaderOPM8) readModuleMetrics(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	opticalOPMInterfaces, err := ekinopsReadOPMMetrics(ctx, m.ports)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}

	mappings, err := ekinopsInterfacesIfIdentifierToSliceIndex(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interface identifier mappings")
	}

	for i, opticalOPMInterface := range opticalOPMInterfaces {
		identifier := m.slotIdentifier + "/" + m.moduleName + "/" + strings.Split(*opticalOPMInterface.Identifier, "(")[0]
		idx, ok := mappings[identifier]
		if !ok {
			return nil, fmt.Errorf("interface for identifier '%s' not found", identifier)
		}
		interfaces[idx].OpticalOPM = &opticalOPMInterfaces[i]
		interfaces[idx].IfAlias = interfaces[idx].OpticalOPM.Label
	}

	return interfaces, nil
}

func ekinopsReadOPMMetrics(ctx context.Context, oids ekinopsOPMOIDs) ([]device.OpticalOPMInterface, error) {
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

	var rxPowerResult []network.SNMPResponse

	if oids.rxPowerOID != "" {
		rxPowerResult, err = con.SNMP.SnmpClient.SNMPWalk(ctx, oids.rxPowerOID)
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk for rx power oid failed")
		}
	}

	var opticalOPMInterfaces []device.OpticalOPMInterface
	for k, identifierResult := range identifierResults {
		var opticalOPMInterface device.OpticalOPMInterface
		identifier, err := identifierResult.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get identifier for optical amplifier interface")
		}
		identifier = strings.Split(identifier, "\n")[0]

		label, err := labelResults[k].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get label for optical amplifier interface")
		}
		label = strings.Split(label, "\n")[0]

		opticalOPMInterface.Identifier = &identifier
		opticalOPMInterface.Label = &label

		if rxPowerResult != nil {
			value, err := rxPowerResult[k].GetValueString()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get tx power for optical amplifier interface")
			}
			valueFloat, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse snmp response to float64")
			}
			if oids.powerTransformFunc != nil {
				valueFloat = oids.powerTransformFunc(valueFloat)
			}

			opticalOPMInterface.RXPower = &valueFloat
		}

		opticalOPMInterfaces = append(opticalOPMInterfaces, opticalOPMInterface)
	}

	// read out channels
	channelsResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oids.channelsOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for identifier oid failed")
	}

	// results to map
	channelValues := make(map[int]map[int]float64)
	for _, channelResult := range channelsResults {
		valueString, err := channelResult.GetValueString()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get snmp response as string (oid: %s)", channelResult.GetOID())
		}
		value, err := strconv.ParseFloat(valueString, 10)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse snmp response to float64 (response: %s)", valueString)
		}
		if oids.powerTransformFunc != nil {
			value = oids.powerTransformFunc(value)
		}

		oidArr := strings.Split(channelResult.GetOID(), ".")
		if oidArr[len(oidArr)-2] == "1" {
			continue
		}

		portIdx, err := strconv.Atoi(oidArr[len(oidArr)-1])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse oid port index to int (index: %s)", oidArr[len(oidArr)-1])
		}

		channelIdx, err := strconv.Atoi(oidArr[len(oidArr)-4])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse oid channel index to int (index: %s)", oidArr[len(oidArr)-4])
		}

		if channelIdx > 776 {
			break
		}

		if _, ok := channelValues[portIdx]; !ok {
			channelValues[portIdx] = make(map[int]float64)
		}
		channelValues[portIdx][channelIdx] = value
	}

	for k := range opticalOPMInterfaces {
		channelNum := 13.0
		for channelIdx := 16; channelIdx <= 776; channelIdx += 8 {
			rxPower := channelValues[k][channelIdx]

			channel := device.OpticalChannel{
				Channel: fmt.Sprintf("C%.2f", channelNum),
				RXPower: &rxPower,
			}

			opticalOPMInterfaces[k].Channels = append(opticalOPMInterfaces[k].Channels, channel)

			channelNum += 0.5
		}
	}

	return opticalOPMInterfaces, nil
}
