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

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	// dot1dBasePortIfIndex
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.2.1.17.1.4.1.2")
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get dot1dBasePortIfIndex, skipping VLANs")
		return interfaces, nil
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

	// jnxExVlanPortStatus
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3")
	if err != nil {
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get jnxExVlanPortStatus, skipping VLANs")
		return interfaces, nil
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
		log.Ctx(ctx).Trace().Err(err).Msg("failed to get jnxExVlanName, skipping VLANs")
		return interfaces, nil
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
