package codecommunicator

import (
	"context"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type timosCommunicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of Nokia devices.
func (c *timosCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	interfaces, err := c.parent.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	// get mapping from every ifIndex to a description
	indexDescriptions, err := getPhysPortDescriptions(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read phys port descriptions")
	}

	// apply description mapping to the default interfaces
	interfaces = normalizeTimosInterfaces(interfaces, indexDescriptions)

	// get all sap interfaces
	sapDescriptionsOID := ".1.3.6.1.4.1.6527.3.1.2.4.3.2.1.5"
	sapDescriptions, err := con.SNMP.SnmpClient.SNMPWalk(ctx, sapDescriptionsOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk failed")
	}

	for _, response := range sapDescriptions {
		special, err := response.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}

		// construct description
		suffix := strings.Split(strings.TrimPrefix(response.GetOID(), sapDescriptionsOID), ".")
		physIndex := suffix[2]
		subID := suffix[3]
		description, ok := indexDescriptions[physIndex]
		if !ok {
			return nil, errors.New("invalid physical index")
		}
		description += ":" + subID
		if special != "" {
			description += " " + special
		}

		// construct index
		subIndex, err := strconv.ParseUint(physIndex+subID, 0, 64)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get index from strings")
		}

		// retrieve admin status
		admin, err := getStatusFromSnmpGet(ctx, ".1.3.6.1.4.1.6527.3.1.2.4.3.2.1.6."+suffix[1]+"."+physIndex+"."+subID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve admin status")
		}

		// retrieve oper status
		oper, err := getStatusFromSnmpGet(ctx, ".1.3.6.1.4.1.6527.3.1.2.4.3.2.1.7."+suffix[1]+"."+physIndex+"."+subID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve oper status")
		}

		// build logical interface
		interfaces = append(interfaces, device.Interface{
			IfIndex:       &subIndex,
			IfDescr:       &description,
			IfAdminStatus: &admin,
			IfOperStatus:  &oper,
		})
	}

	return interfaces, nil
}

// getPhysPortDescriptions returns a mapping from every ifIndex to a description.
// This description is different and shorter than the ifDescription.
func getPhysPortDescriptions(ctx context.Context) (map[string]string, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	physPortsOID := ".1.3.6.1.4.1.6527.3.1.2.2.4.2.1.6.1"
	physPorts, err := con.SNMP.SnmpClient.SNMPWalk(ctx, physPortsOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk failed")
	}

	indexDescriptions := make(map[string]string)

	for _, response := range physPorts {
		description, err := response.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get string value")
		}
		index := strings.TrimPrefix(response.GetOID(), physPortsOID+".")
		indexDescriptions[index] = description
	}
	return indexDescriptions, nil
}

// getCounterFromSnmpGet returns the snmpget value at the given oid as uint64 counter.
func getCounterFromSnmpGet(ctx context.Context, oid string) (uint64, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return 0, errors.New("no device connection available")
	}

	res, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
	if err != nil || len(res) != 1 {
		return 0, errors.Wrap(err, "snmpget failed")
	}
	resString, err := res[0].GetValueString()
	if err != nil {
		return 0, errors.Wrap(err, "couldn't parse snmp response")
	}
	resCounter, err := strconv.ParseUint(resString, 0, 64)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't parse snmp response")
	}
	return resCounter, nil
}

// getStatusFromSnmpGet returns the snmpget value at the given oid as device.Status.
func getStatusFromSnmpGet(ctx context.Context, oid string) (device.Status, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return "", errors.New("no device connection available")
	}

	res, err := con.SNMP.SnmpClient.SNMPGet(ctx, oid)
	if err != nil || len(res) != 1 {
		return "", errors.Wrap(err, "snmpget failed")
	}
	resString, err := res[0].GetValueString()
	if err != nil {
		return "", errors.Wrap(err, "couldn't parse snmp response")
	}
	resInt, err := strconv.Atoi(resString)
	if err != nil {
		return "", errors.Wrap(err, "couldn't parse snmp response")
	}
	resStatus, err := device.GetStatus(resInt)
	if err != nil {
		return "", errors.Wrap(err, "couldn't get status from snmp response")
	}
	return resStatus, nil
}

// normalizeTimosInterfaces applies the description mapping to the given interfaces.
func normalizeTimosInterfaces(interfaces []device.Interface, descriptions map[string]string) []device.Interface {
	for _, interf := range interfaces {
		descr, ok := descriptions[strconv.FormatUint(*interf.IfIndex, 10)]
		if !ok {
			continue
		}
		*interf.IfDescr = descr
		*interf.IfName = descr
	}

	return interfaces
}
