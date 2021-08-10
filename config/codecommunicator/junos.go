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

type junosCommunicator struct {
	codeCommunicator
}

func (c *junosCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	interfaces, err := c.deviceClass.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	interfacesWithVLANs, err := juniperAddVLANsNonELS(ctx, interfaces)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("getting juniper VLANs for non ELS devices failed, trying for ELS devices")
		interfacesWithVLANs, err = juniperAddVLANsELS(ctx, interfaces)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("getting juniper VLANs for ELS devices failed, skipping VLANs")
			interfacesWithVLANs = interfaces
		}
	}

	return interfacesWithVLANs, nil
}

func juniperAddVLANsELS(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	// jnxL2aldVlanFdbId
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.48.1.3.1.1.5")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxL2aldVlanFdbId")
	}

	vlanIndexFilterID := make(map[string]string)
	for _, response := range res {
		filterID, err := response.GetValueString()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid, ".")

		vlanIndexFilterID[oidSplit[len(oidSplit)-1]] = filterID
	}

	// jnxL2aldVlanName
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.48.1.3.1.1.2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxL2aldVlanName")
	}

	filterIDVLAN := make(map[string]device.VLAN)
	for _, response := range res {
		name, err := response.GetValueString()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid, ".")
		filterID := vlanIndexFilterID[oidSplit[len(oidSplit)-1]]

		filterIDVLAN[filterID] = device.VLAN{
			Name: name,
		}
	}

	portIfIndex, err := juniperGetPortIfIndexMapping(ctx)
	if err != nil {
		return nil, err
	}

	// dot1qTpFdbPort
	dot1qTpFdbPort := "1.3.6.1.2.1.17.7.1.2.2.1.2"
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, dot1qTpFdbPort)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dot1qTpFdbPort")
	}

	ifIndexFilterIDs := make(map[string][]string)
out:
	for _, response := range res {
		port, err := response.GetValueString()
		if err != nil {
			return nil, err
		}

		oid := strings.TrimPrefix(response.GetOID(), ".")
		oidSplit := strings.Split(strings.TrimPrefix(strings.TrimPrefix(oid, dot1qTpFdbPort), "."), ".")
		ifIndex := portIfIndex[port]

		for _, filterID := range ifIndexFilterIDs[ifIndex] {
			if filterID == oidSplit[0] {
				continue out
			}
		}
		ifIndexFilterIDs[ifIndex] = append(ifIndexFilterIDs[ifIndex], oidSplit[0])
	}

	for i, interf := range interfaces {
		if interf.IfIndex != nil {
			if filterIDs, ok := ifIndexFilterIDs[fmt.Sprint(*interf.IfIndex)]; ok {
				for _, filterID := range filterIDs {
					if vlan, ok := filterIDVLAN[filterID]; ok {
						if interfaces[i].VLAN == nil {
							interfaces[i].VLAN = &device.VLANInformation{}
						}
						interfaces[i].VLAN.VLANs = append(interfaces[i].VLAN.VLANs, vlan)
					}
				}
			}
		}
	}

	return interfaces, nil
}

func juniperAddVLANsNonELS(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	// jnxExVlanPortStatus
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxExVlanPortStatus")
	}

	portIfIndex, err := juniperGetPortIfIndexMapping(ctx)
	if err != nil {
		return nil, err
	}

	vlanIndexVLAN := make(map[string]device.VLAN)
	ifIndexVLANIndices := make(map[string][]string)
	for _, response := range res {
		status, err := response.GetValueString()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid, ".")

		ifIndex := portIfIndex[oidSplit[len(oidSplit)-1]]
		ifIndexVLANIndices[ifIndex] = append(ifIndexVLANIndices[ifIndex], oidSplit[len(oidSplit)-2])
		vlanIndexVLAN[oidSplit[len(oidSplit)-2]] = device.VLAN{
			Status: &status,
		}
	}

	// jnxExVlanName
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.40.1.5.1.5.1.2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxExVlanName")
	}

	for _, response := range res {
		name, err := response.GetValueString()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid, ".")

		if vlan, ok := vlanIndexVLAN[oidSplit[len(oidSplit)-1]]; ok {
			vlan.Name = name
			vlanIndexVLAN[oidSplit[len(oidSplit)-1]] = vlan
		}
	}

	for i, interf := range interfaces {
		if interf.IfIndex != nil {
			if vlanIndices, ok := ifIndexVLANIndices[fmt.Sprint(*interf.IfIndex)]; ok {
				for _, vlanIndex := range vlanIndices {
					if vlan, ok := vlanIndexVLAN[vlanIndex]; ok {
						if interfaces[i].VLAN == nil {
							interfaces[i].VLAN = &device.VLANInformation{}
						}
						interfaces[i].VLAN.VLANs = append(interfaces[i].VLAN.VLANs, vlan)
					}
				}
			}
		}
	}

	return interfaces, nil
}

func juniperGetPortIfIndexMapping(ctx context.Context) (map[string]string, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	// dot1dBasePortIfIndex
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.17.1.4.1.2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dot1dBasePortIfIndex")
	}

	portIfIndex := make(map[string]string)
	for _, response := range res {
		ifIndex, err := response.GetValueString()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid, ".")

		portIfIndex[oidSplit[len(oidSplit)-1]] = ifIndex
	}

	return portIfIndex, nil
}
