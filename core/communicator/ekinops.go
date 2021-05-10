package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

type ekinopsCommunicator struct {
	baseCommunicator
}

// GetInterfaces returns the interfaces of ekinops devices.
func (c *ekinopsCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	interfaces, err := c.GetIfTable(ctx)
	if err != nil {
		return nil, err
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
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
		slotIdentifier, err := slotResult.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get snmp result as string")
		}

		module, err := moduleResults[k].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get module result as string")
		}

		moduleReader, err := ekinopsGetModuleReader(slotIdentifier, module)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msgf("no information for reading out ekinops module '%s' available", module)
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

	return normalizeEkinopsInterfaces(interfaces)
}

// GetIfTable returns the ifTable of ekinops devices.
// For ekinops devices, only a few interface values are required.
func (c *ekinopsCommunicator) GetIfTable(ctx context.Context) ([]device.Interface, error) {
	if genericDeviceClass.components.interfaces.Values == nil {
		return nil, errors.New("ifTable information is empty")
	}

	reader := *genericDeviceClass.components.interfaces.Values.(*snmpGroupPropertyReader)
	oids := make(deviceClassOIDs)

	regex, err := regexp.Compile("(ifIndex|ifDescr|ifType|ifName|ifAdminStatus|ifOperStatus|ifPhysAddress)")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build regex")
	}

	for oid, value := range reader.oids {
		if regex.MatchString(oid) {
			oids[oid] = value
		}
	}
	reader.oids = oids

	interfacesRaw, err := reader.getProperty(ctx)
	if err != nil {
		return nil, err
	}

	var interfaces []device.Interface

	err = interfacesRaw.Decode(&interfaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode raw interfaces into interface structs")
	}

	return interfaces, nil
}

func ekinopsInterfacesIfIdentifierToSliceIndex(interfaces []device.Interface) (map[string]int, error) {
	m := make(map[string]int)
	for k, interf := range interfaces {
		if interf.IfName == nil {
			return nil, fmt.Errorf("no ifName set for interface ifIndex: `%d`", *interf.IfIndex)
		}
		identifier := strings.Split(strings.Join(strings.Split(*interf.IfName, "/")[2:], "/"), "(")[0]

		if _, ok := m[identifier]; ok {
			return nil, fmt.Errorf("interface identifier `%s` exists multiple times", *interf.IfName)
		}

		m[identifier] = k
	}
	return m, nil
}

func normalizeEkinopsInterfaces(interfaces []device.Interface) ([]device.Interface, error) {
	var res []device.Interface

	for _, interf := range interfaces {
		if interf.IfDescr == nil {
			return nil, fmt.Errorf("no IfDescr set for interface ifIndex: `%d`", *interf.IfIndex)
		}

		slotNumber := strings.Split(*interf.IfName, "/")[2]
		moduleName := strings.Split(*interf.IfDescr, "/")[3]

		// change ifType of ports of slots > 0 to "opticalChannel" if ifType equals "other", but not OPM8 interfaces
		if slotNumber != "0" && interf.IfType != nil && *interf.IfType == "other" && moduleName != "PM_OPM8" {
			opticalChannel := "opticalChannel"
			interf.IfType = &opticalChannel
		}

		// change subType of OPM8 ports
		if moduleName == "PM_OPM8" {
			subType := "channelMonitoring"
			interf.SubType = &subType
		}

		// change ifDescr and ifName of every interface
		// they no longer contain Ekinops/C...
		// and are now in the form <slot>-[Description]
		*interf.IfDescr = slotNumber + "-" + strings.Split(strings.Split(*interf.IfDescr, "/")[4], "(")[0]
		interf.IfName = interf.IfDescr

		// remove every port on slot 0 starting with "0-FE_"
		if slotNumber != "0" || !strings.HasPrefix(*interf.IfName, "0-FE_") {
			res = append(res, interf)
		}
	}

	return res, nil
}
