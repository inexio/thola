package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type timosSASCommunicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of a Nokia SAS-T device.
func (c *timosSASCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	interfaces, err := c.parent.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	// get all sap interfaces
	sapDescriptionsOID := network.OID(".1.3.6.1.4.1.6527.3.1.2.4.3.2.1.5")
	sapDescriptions, err := con.SNMP.SnmpClient.SNMPWalk(ctx, sapDescriptionsOID)
	if err != nil {
		return nil, errors.Wrap(err, "snmpwalk failed")
	}

	for _, response := range sapDescriptions {
		// construct description
		suffix := strings.Split(strings.TrimPrefix(response.GetOID().String(), sapDescriptionsOID.String()), ".")
		physIndex := suffix[2]
		subID := suffix[3]

		// construct index
		subIndex, err := strconv.ParseUint(physIndex+subID, 0, 64)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get index from strings")
		}

		// search sap interface that matches given subIndex
		i, err := getInterfaceBySubIndex(subIndex, interfaces)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get interface from index")
		}

		// retrieve inbound
		inbound, err := getCounterFromSnmpGet(ctx, network.OID(".1.3.6.1.4.1.6527.6.2.2.2.8.1.1.1.4.").AddIndex(suffix[1]+"."+physIndex+"."+subID))
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve inbound counter")
		}

		// retrieve outbound
		outbound, err := getCounterFromSnmpGet(ctx, network.OID(".1.3.6.1.4.1.6527.6.2.2.2.8.1.1.1.6.").AddIndex(suffix[1]+"."+physIndex+"."+subID))
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve outbound counter")
		}

		// append the sap struct to the interface
		interfaces[i].SAP = &device.SAPInterface{
			Inbound:  &inbound,
			Outbound: &outbound,
		}
	}

	return filterInterfaces(interfaces, filter)
}

// getInterfaceBySubIndex returns the index of the interface that has the given index.
// The returned index is the index of the array, not the IfIndex.
func getInterfaceBySubIndex(subIndex uint64, interfaces []device.Interface) (int, error) {
	for index, iface := range interfaces {
		if iface.IfIndex != nil && *iface.IfIndex == subIndex {
			return index, nil
		}
	}
	return 0, errors.New("no interface with given index found")
}
