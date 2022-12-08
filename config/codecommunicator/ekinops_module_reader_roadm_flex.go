package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type ekinopsModuleReaderRoadmFlex struct {
	ekinopsModuleData
	ports ekinopsRoadmFlexOIDs
}

type ekinopsRoadmFlexOIDs struct {
	identifierOID network.OID
	labelOID      network.OID
	rxPowerOID    network.OID
	channelsOid   network.OID

	powerTransformFunc ekinopsPowerTransformFunc
}

func (m *ekinopsModuleReaderRoadmFlex) readModuleMetrics(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	opticalRoadmFlexInterfaces, err := ekinopsReadRoadmFlexMetrics(ctx, m.ports)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out metrics for booster ports")
	}

	mappings, err := ekinopsInterfacesIfIdentifierToSliceIndex(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interface identifier mappings")
	}

	for i, opticalRoadmFlexInterface := range opticalRoadmFlexInterfaces {
		identifier := m.slotIdentifier + "/" + m.moduleName + "/" + strings.Split(*opticalRoadmFlexInterface.Identifier, "(")[0]
		idx, ok := mappings[identifier]
		if !ok {
			log.Ctx(ctx).Debug().Msgf("interface for identifier '%s' not found", identifier)
			continue
		}
		interfaces[idx].OpticalRoadmFlex = &opticalRoadmFlexInterfaces[i]
		interfaces[idx].IfAlias = interfaces[idx].OpticalRoadmFlex.Label
	}

	return interfaces, nil
}

func ekinopsReadRoadmFlexMetrics(ctx context.Context, oids ekinopsRoadmFlexOIDs) ([]device.OpticalRoadmFlexInterface, error) {
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

	var rxPowerResults []network.SNMPResponse

	if oids.rxPowerOID != "" {
		rxPowerResults, err = con.SNMP.SnmpClient.SNMPWalk(ctx, oids.rxPowerOID)
		if err != nil {
			return nil, errors.Wrap(err, "snmpwalk for rx power oid failed")
		}
	}

	var opticalRoadmFlexInterfaces []device.OpticalRoadmFlexInterface
	for k, identifierResult := range identifierResults {
		var opticalRoadmFlexInterface device.OpticalRoadmFlexInterface
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

		opticalRoadmFlexInterface.Identifier = &identifierString
		opticalRoadmFlexInterface.Label = &labelString

		if rxPowerResults != nil {
			value, err := rxPowerResults[k].GetValue()
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

			opticalRoadmFlexInterface.RXPower = &valueFloat
		}

		opticalRoadmFlexInterfaces = append(opticalRoadmFlexInterfaces, opticalRoadmFlexInterface)
	}

	// read out channels
	channelsResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, oids.channelsOid)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk for identifier oid failed")
	}

	// results to map
	channelValuesIn := make(map[int]float64)
	channelValuesOut := make(map[int]float64)
	for _, channelResult := range channelsResults {
		val, err := channelResult.GetValue()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get snmp response as string (oid: %s)", channelResult.GetOID())
		}
		value, err := val.Float64()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse snmp response to float64 (response: %s)", val)
		}
		if oids.powerTransformFunc != nil {
			value = oids.powerTransformFunc(value)
		}

		oidArr := strings.Split(channelResult.GetOID().String(), ".")

		channelIdx, err := strconv.Atoi(oidArr[len(oidArr)-4])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse oid channel index to int (index: %s)", oidArr[len(oidArr)-4])
		}

		if oidArr[len(oidArr)-2] == "1" || channelIdx < 65 || channelIdx > 157 || channelIdx%2 == 0 {
			continue
		}

		portIdx, err := strconv.Atoi(oidArr[len(oidArr)-1])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse oid port index to int (index: %s)", oidArr[len(oidArr)-1])
		}
		channelNum := (channelIdx-65)/2 + 14

		if portIdx == 0 {
			channelValuesIn[channelNum] = value
		} else if portIdx == 1 {
			channelValuesOut[channelNum] = value
		} else {
			log.Ctx(ctx).Warn().Msgf("unexpected channel number %d for oid %s", portIdx, channelResult.GetOID().String())
		}
	}

	for channelNum := 14; channelNum <= 60; channelNum += 1 {
		channelName := fmt.Sprintf("C%d", channelNum)
		rxPowerIn := channelValuesIn[channelNum]
		rxPowerOut := channelValuesOut[channelNum]

		channelIn := device.OpticalChannel{
			Channel: &channelName,
			RXPower: &rxPowerIn,
		}
		channelOut := device.OpticalChannel{
			Channel: &channelName,
			RXPower: &rxPowerOut,
		}
		opticalRoadmFlexInterfaces[0].Channels = append(opticalRoadmFlexInterfaces[0].Channels, channelIn)
		opticalRoadmFlexInterfaces[1].Channels = append(opticalRoadmFlexInterfaces[1].Channels, channelOut)
	}

	return opticalRoadmFlexInterfaces, nil
}
