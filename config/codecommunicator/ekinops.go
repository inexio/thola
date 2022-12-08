package codecommunicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

type ekinopsCommunicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of ekinops devices.
func (c *ekinopsCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	con.SNMP.SnmpClient.UseCache(false)

	interfaces, err := c.deviceClass.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	//get used slots
	slotResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.20044.7.8.1.1.2")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to snmpwalk")
	}

	//get used modules
	moduleResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.20044.7.8.1.1.3")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to snmpwalk")
	}

	var moduleReaders []ekinopsModuleReader

	for k, slotResult := range slotResults {
		slotIdentifier, err := slotResult.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get snmp result as string")
		}

		module, err := moduleResults[k].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get module result as string")
		}

		moduleReader, err := ekinopsGetModuleReader(slotIdentifier.String(), module.String())
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msgf("no information for reading out ekinops module '%s' available", module)
			continue
		}
		moduleReaders = append(moduleReaders, moduleReader)
	}

	for _, moduleReader := range moduleReaders {
		interfaces, err = moduleReader.readModuleMetrics(ctx, interfaces)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read out module specific metrics for module '%s' (slot: %s)", moduleReader.getModuleName(), moduleReader.getSlotIdentifier())
		}
	}

	interfaces, err = c.normalizeInterfaces(interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to normalize interfaces")
	}

	return filterInterfaces(ctx, interfaces, filter)
}

func ekinopsInterfacesIfIdentifierToSliceIndex(interfaces []device.Interface) (map[string]int, error) {
	m := make(map[string]int)
	interfaceRegex, _ := regexp.Compile("([0-9]+)/(PM_[^/]+|MGNT)/([^\\(]+)")
	for k, interf := range interfaces {
		if interf.IfName == nil {
			return nil, fmt.Errorf("no ifName set for interface ifIndex: `%d`", *interf.IfIndex)
		}
		match := interfaceRegex.FindStringSubmatch(*interf.IfName)
		identifier := match[1] + "/" + match[2] + "/" + match[3]

		if _, ok := m[identifier]; ok {
			return nil, fmt.Errorf("interface identifier `%s` exists multiple times", *interf.IfName)
		}

		m[identifier] = k
	}
	return m, nil
}

func (c *ekinopsCommunicator) normalizeInterfaces(interfaces []device.Interface) ([]device.Interface, error) {
	var res []device.Interface

	// if_descr is for example: EKINOPS/C600HC/20/PM_OPM8/OPM-4(S14_from_Oerel)
	//                         EKINOPS/R1/Su1/Sl8/PM_ROADM-FLEX-H10M/WSS_Line_In(WSS_LINE_IN     )
	//                         EKINOPS/R1/Su1/Sl0/MGNT/FE_1
	interfaceRegex, _ := regexp.Compile("([0-9]+)/(PM_[^/]+|MGNT)/([^\\(]+)")

	for _, interf := range interfaces {
		if interf.IfDescr == nil {
			return nil, fmt.Errorf("no IfDescr set for interface ifIndex: `%d`", *interf.IfIndex)
		}

		match := interfaceRegex.FindStringSubmatch(*interf.IfDescr)
		log.Debug().Msgf("found slot %s, module %s, port %s", match[1], match[2], match[3])

		slotNumber := match[1]
		moduleName := match[2]
		portName := match[3]

		// change ifType of ports of slots > 0 to "opticalChannel" if ifType equals "other", but not OPM8 interfaces
		if slotNumber != "0" && interf.IfType != nil && *interf.IfType == "other" && moduleName != "PM_OPM8" {
			opticalChannel := "opticalChannel"
			interf.IfType = &opticalChannel
		}

		// change subType of OPM8 ports
		if moduleName == "PM_OPM8" || moduleName == "PM_ROADM-FLEX-H4M" || moduleName == "PM_ROADM-FLEX-H10M" {
			subType := "channelMonitoring"
			interf.SubType = &subType
		}

		// change ifDescr and ifName of every interface
		// they no longer contain Ekinops/C...
		// and are now in the form <slot>-[Description]
		*interf.IfDescr = slotNumber + "-" + portName
		interf.IfName = interf.IfDescr

		// remove every port on slot 0 starting with "0-FE_"
		if slotNumber != "0" || !strings.HasPrefix(*interf.IfName, "0-FE_") {
			res = append(res, interf)
		}
	}

	return res, nil
}
