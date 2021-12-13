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

type junosCommunicator struct {
	codeCommunicator
}

func (c *junosCommunicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	interfaces, err := c.deviceClass.GetInterfaces(ctx, filter...)
	if err != nil {
		return nil, err
	}

	if groupproperty.CheckValueFiltersMatch(filter, []string{"vlan"}) {
		log.Ctx(ctx).Debug().Msg("filter matched on 'vlan', skipping junos vlan values")
		return interfaces, nil
	}
	log.Ctx(ctx).Debug().Msg("reading junos vlan values")

	interfacesWithVLANs, err := c.addVLANsNonELS(ctx, interfaces)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("getting juniper VLANs for non ELS devices failed, trying for ELS devices")
		interfacesWithVLANs, err = c.addVLANsELS(ctx, interfaces)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("getting juniper VLANs for ELS devices failed, skipping VLANs")
			interfacesWithVLANs = interfaces
		}
	}

	return filterInterfaces(ctx, interfacesWithVLANs, filter)
}

func (c *junosCommunicator) addVLANsELS(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
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
		filterID, err := response.GetValue()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid.String(), ".")

		vlanIndexFilterID[oidSplit[len(oidSplit)-1]] = filterID.String()
	}

	// jnxL2aldVlanName
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.48.1.3.1.1.2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxL2aldVlanName")
	}

	filterIDVLAN := make(map[string]device.VLAN)
	for _, response := range res {
		name, err := response.GetValue()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid.String(), ".")
		filterID := vlanIndexFilterID[oidSplit[len(oidSplit)-1]]
		nameString := name.String()

		filterIDVLAN[filterID] = device.VLAN{
			Name: &nameString,
		}
	}

	portIfIndex, err := c.getPortIfIndexMapping(ctx)
	if err != nil {
		return nil, err
	}

	// dot1qTpFdbPort
	dot1qTpFdbPort := network.OID("1.3.6.1.2.1.17.7.1.2.2.1.2")
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, dot1qTpFdbPort)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dot1qTpFdbPort")
	}

	ifIndexFilterIDs := make(map[string][]string)
out:
	for _, response := range res {
		port, err := response.GetValue()
		if err != nil {
			return nil, err
		}

		oid := strings.TrimPrefix(response.GetOID().String(), ".")
		oidSplit := strings.Split(strings.TrimPrefix(strings.TrimPrefix(oid, dot1qTpFdbPort.String()), "."), ".")
		ifIndex := portIfIndex[port.String()]

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

func (c *junosCommunicator) addVLANsNonELS(ctx context.Context, interfaces []device.Interface) ([]device.Interface, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("snmp client is empty")
	}

	// jnxExVlanPortStatus
	res, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.40.1.5.1.7.1.3")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxExVlanPortStatus")
	}

	portIfIndex, err := c.getPortIfIndexMapping(ctx)
	if err != nil {
		return nil, err
	}

	vlanIndexVLAN := make(map[string]device.VLAN)
	ifIndexVLANIndices := make(map[string][]string)
	for _, response := range res {
		status, err := response.GetValue()
		if err != nil {
			return nil, err
		}
		statusString := status.String()

		oid := response.GetOID()
		oidSplit := strings.Split(oid.String(), ".")

		ifIndex := portIfIndex[oidSplit[len(oidSplit)-1]]
		ifIndexVLANIndices[ifIndex] = append(ifIndexVLANIndices[ifIndex], oidSplit[len(oidSplit)-2])
		vlanIndexVLAN[oidSplit[len(oidSplit)-2]] = device.VLAN{
			Status: &statusString,
		}
	}

	// jnxExVlanName
	res, err = con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2636.3.40.1.5.1.5.1.2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get jnxExVlanName")
	}

	for _, response := range res {
		name, err := response.GetValue()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid.String(), ".")

		if vlan, ok := vlanIndexVLAN[oidSplit[len(oidSplit)-1]]; ok {
			vlanName := name.String()
			vlan.Name = &vlanName
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

func (c *junosCommunicator) getPortIfIndexMapping(ctx context.Context) (map[string]string, error) {
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
		ifIndex, err := response.GetValue()
		if err != nil {
			return nil, err
		}

		oid := response.GetOID()
		oidSplit := strings.Split(oid.String(), ".")

		portIfIndex[oidSplit[len(oidSplit)-1]] = ifIndex.String()
	}

	return portIfIndex, nil
}

func (c *junosCommunicator) GetCPUComponentCPULoad(ctx context.Context) ([]device.CPU, error) {
	indices, err := c.getRoutingEngineIndices(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get routing indices")
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	jnxOperatingCPUOID := network.OID(".1.3.6.1.4.1.2636.3.1.13.1.8")
	var cpus []device.CPU
	for i, index := range indices {
		response, err := con.SNMP.SnmpClient.SNMPGet(ctx, jnxOperatingCPUOID.AddIndex(index.index))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get CPU load")
		} else if len(response) != 1 {
			return nil, errors.New("invalid cpu load result")
		}

		res, err := response[0].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get string value of cpu load")
		}

		load, err := res.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse cpu load")
		}

		cpus = append(cpus, device.CPU{
			Label: &indices[i].label,
			Load:  &load,
		})
	}

	spuCpus, err := c.getSPUCPUs(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to read out SPU CPU load")
	} else {
		cpus = append(cpus, spuCpus...)
	}

	return cpus, nil
}

type indexAndLabel struct {
	index string
	label string
}

func (c *junosCommunicator) getRoutingEngineIndices(ctx context.Context) ([]indexAndLabel, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	jnxOperatingDescrOID := network.OID(".1.3.6.1.4.1.2636.3.1.13.1.5")
	jnxOperatingDescr, err := con.SNMP.SnmpClient.SNMPWalk(ctx, jnxOperatingDescrOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get 'jnxOperatingDescrOID'")
	}

	var indices []indexAndLabel
	for _, response := range jnxOperatingDescr {
		res, err := response.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get string value of snmp response")
		}

		if ok, err = regexp.MatchString("(?i)engine", res.String()); err == nil && ok {
			indices = append(indices, indexAndLabel{
				index: strings.TrimPrefix(response.GetOID().String(), jnxOperatingDescrOID.String()),
				label: res.String(),
			})
		}
	}

	return indices, nil
}

func (c *junosCommunicator) getSPUCPUs(ctx context.Context) ([]device.CPU, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	indexDescr, err := c.getSPUIndices(ctx)
	if err != nil {
		return nil, err
	}

	jnxJsSPUMonitoringCPUUsageOID := network.OID(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.4")
	jnxJsSPUMonitoringCPUUsage, err := con.SNMP.SnmpClient.SNMPWalk(ctx, jnxJsSPUMonitoringCPUUsageOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get 'jnxJsSPUMonitoringCPUUsage'")
	}

	var cpus []device.CPU
	for _, load := range jnxJsSPUMonitoringCPUUsage {
		res, err := load.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get string value of snmp response")
		}
		resParsed, err := res.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse snmp response")
		}
		cpu := device.CPU{Load: &resParsed}
		if descr, ok := indexDescr[strings.TrimPrefix(load.GetOID().String(), jnxJsSPUMonitoringCPUUsageOID.String())]; ok {
			label := "spu_" + descr
			cpu.Label = &label
		}
		cpus = append(cpus, cpu)
	}

	return cpus, nil
}

func (c *junosCommunicator) getSPUIndices(ctx context.Context) (map[string]string, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	jnxJsSPUMonitoringNodeDescrOID := network.OID(".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.11")
	jnxJsSPUMonitoringNodeDescr, err := con.SNMP.SnmpClient.SNMPWalk(ctx, jnxJsSPUMonitoringNodeDescrOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get 'jnxJsSPUMonitoringNodeDescr'")
	}

	indexDescr := make(map[string]string)
	for _, descr := range jnxJsSPUMonitoringNodeDescr {
		res, err := descr.GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get string value of snmp response")
		}
		indexDescr[strings.TrimPrefix(descr.GetOID().String(), jnxJsSPUMonitoringNodeDescrOID.String())] = res.String()
	}

	return indexDescr, nil
}

func (c *junosCommunicator) GetMemoryComponentMemoryUsage(ctx context.Context) ([]device.MemoryPool, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	var pools []device.MemoryPool

	// kernel memory used
	kernelMemUsedRes, err := con.SNMP.SnmpClient.SNMPGet(ctx, "1.3.6.1.4.1.2636.3.1.16.0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read out kernel memory usage")
	}
	if len(kernelMemUsedRes) > 0 {
		val, err := kernelMemUsedRes[0].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert kernel memory usage to string")
		}
		f, err := val.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert kernel memory usage to float64")
		}
		label := "kernel"
		pools = append(pools, device.MemoryPool{
			Label: &label,
			Usage: &f,
		})
	}

	// engines
	indices, err := c.getRoutingEngineIndices(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get routing engine indices")
	}
	for i, index := range indices {
		response, err := con.SNMP.SnmpClient.SNMPGet(ctx, network.OID(".1.3.6.1.4.1.2636.3.1.13.1.11").AddIndex(index.index))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get memory usage")
		} else if len(response) != 1 {
			return nil, errors.New("invalid memory usage result")
		}

		res, err := response[0].GetValue()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get string value of memory usage")
		}

		usage, err := res.Float64()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse memory usage")
		}

		pools = append(pools, device.MemoryPool{
			Label: &indices[i].label,
			Usage: &usage,
		})
	}

	// spu
	spuIndexDescr, err := c.getSPUIndices(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to read out SPU memory usage")
	} else {
		spuOID := ".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.5"
		spuUsages, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.2636.3.39.1.12.1.1.1.5")
		if err != nil {
			return nil, errors.Wrap(err, "failed to get spu memory usages")
		}

		for _, res := range spuUsages {
			resVal, err := res.GetValue()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get string value of snmp response")
			}
			resParsed, err := resVal.Float64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse snmp response")
			}
			pool := device.MemoryPool{Usage: &resParsed}
			if descr, ok := spuIndexDescr[strings.TrimPrefix(res.GetOID().String(), spuOID)]; ok {
				label := "spu_" + descr
				pool.Label = &label
			}
			pools = append(pools, pool)
		}
	}

	return pools, nil
}
