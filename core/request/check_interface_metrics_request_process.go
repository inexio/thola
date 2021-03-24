// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/parser"
	"github.com/pkg/errors"
	"regexp"
)

type interfaceCheckOutput struct {
	IfIndex       string `json:"ifIndex"`
	IfDescr       string `json:"ifDescr"`
	IfType        string `json:"ifType"`
	IfName        string `json:"ifName"`
	IfAlias       string `json:"ifAlias"`
	IfPhysAddress string `json:"ifPhysAddress"`
	IfAdminStatus string `json:"ifAdminStatus"`
	IfOperStatus  string `json:"ifOperStatus"`

	SubType string `json:"subType"`
}

func (r *CheckInterfaceMetricsRequest) process(ctx context.Context) (Response, error) {
	r.init()

	readInterfacesResponse, err := r.getData(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read interfaces request", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if r.PrintInterfaces {
		var interfaces []interfaceCheckOutput
		for _, interf := range readInterfacesResponse.Interfaces {
			x := interfaceCheckOutput{}
			if interf.IfIndex != nil {
				x.IfIndex = fmt.Sprint(*interf.IfIndex)
			}
			if interf.IfDescr != nil {
				x.IfDescr = *interf.IfDescr
			}
			if interf.IfName != nil {
				x.IfName = *interf.IfName
			}
			if interf.IfType != nil {
				x.IfType = *interf.IfType
			}
			if interf.IfAlias != nil {
				x.IfAlias = *interf.IfAlias
			}
			if interf.IfPhysAddress != nil {
				x.IfPhysAddress = *interf.IfPhysAddress
			}
			if interf.IfAdminStatus != nil {
				x.IfAdminStatus = string(*interf.IfAdminStatus)
			}
			if interf.IfOperStatus != nil {
				x.IfOperStatus = string(*interf.IfOperStatus)
			}
			if interf.SubType != nil {
				x.SubType = *interf.SubType
			}
			interfaces = append(interfaces, x)
		}
		output, err := parser.Parse(interfaces, "json")
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while marshalling output", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		r.mon.UpdateStatus(monitoringplugin.OK, string(output))
	}

	err = addCheckInterfacePerformanceData(readInterfacesResponse.Interfaces, r.mon)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data", true) {
		r.mon.PrintPerformanceData(false)
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}

func (r *CheckInterfaceMetricsRequest) getData(ctx context.Context) (*ReadInterfacesResponse, error) {
	readInterfacesRequest := ReadInterfacesRequest{ReadRequest{r.BaseRequest}}
	response, err := readInterfacesRequest.process(ctx)
	if err != nil {
		return nil, err
	}

	readInterfacesResponse := response.(*ReadInterfacesResponse)

	var filterIndices []int
out:
	for i, interf := range readInterfacesResponse.Interfaces {
		for _, filter := range r.IfTypeFilter {
			if interf.IfType != nil && *interf.IfType == filter {
				filterIndices = append(filterIndices, i)
				continue out
			}
		}
		for _, filter := range r.IfNameFilter {
			if interf.IfName != nil {
				matched, err := regexp.MatchString(filter, *interf.IfName)
				if err != nil {
					return nil, errors.Wrap(err, "ifName filter regex match failed")
				}
				if matched {
					filterIndices = append(filterIndices, i)
					continue out
				}
			}
		}
	}

	readInterfacesResponse.Interfaces = filterInterfaces(readInterfacesResponse.Interfaces, filterIndices, 0)

	return readInterfacesResponse, nil
}

func filterInterfaces(interfaces []device.Interface, toRemove []int, alreadyRemoved int) []device.Interface {
	if len(toRemove) == 0 {
		return interfaces
	}
	return append(interfaces[:toRemove[0]-alreadyRemoved], filterInterfaces(interfaces[toRemove[0]+1-alreadyRemoved:], toRemove[1:], toRemove[0]+1)...)
}

func addCheckInterfacePerformanceData(interfaces []device.Interface, r *monitoringplugin.Response) error {
	ifDescriptions := make(map[string]*device.Interface)

	// if the device has multiple interfaces with the same ifDescr, the ifDescr will be modified and the ifIndex will be attached
	// otherwise, the monitoring plugin will throw an error because of duplicate labels
	for i, origInterf := range interfaces {
		if origInterf.IfDescr != nil {
			if interf, ok := ifDescriptions[*origInterf.IfDescr]; ok {
				if interf != nil {
					if interf.IfIndex == nil {
						return errors.New("interface does not have an ifIndex, but ifDescr is a duplicate")
					}
					ifDescr := *interf.IfDescr + " " + fmt.Sprint(*interf.IfIndex)
					interf.IfDescr = &ifDescr
					ifDescriptions[*origInterf.IfDescr] = nil
				}
				if origInterf.IfIndex == nil {
					return errors.New("interface does not have an ifIndex, but ifDescr is a duplicate")
				}
				ifDescr := *origInterf.IfDescr + " " + fmt.Sprint(*origInterf.IfIndex)
				interfaces[i].IfDescr = &ifDescr
			} else {
				ifDescriptions[*origInterf.IfDescr] = &interfaces[i]
			}
		} else {
			if interfaces[i].IfIndex == nil {
				return errors.New("interface does not have an ifDescription and ifIndex")
			}
			x := fmt.Sprint(*interfaces[i].IfIndex)
			interfaces[i].IfDescr = &x
		}
	}

	for _, i := range interfaces {
		//error_counter_in
		if i.IfInErrors != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_in", *i.IfInErrors).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//error_counter_out
		if i.IfOutErrors != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_out", *i.IfOutErrors).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_discard_in
		if i.IfInDiscards != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_discard_in", *i.IfInDiscards).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_discard_out
		if i.IfOutDiscards != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_discard_out", *i.IfOutDiscards).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//interface_admin_status
		if i.IfAdminStatus != nil {
			value, err := i.IfAdminStatus.ToStatusCode()
			if err != nil {
				return errors.Wrap(err, "failed to convert admin status")
			}
			err = r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_admin_status", value).SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//interface_oper_status
		if i.IfOperStatus != nil {
			value, err := i.IfOperStatus.ToStatusCode()
			if err != nil {
				return errors.Wrap(err, "failed to convert oper status")
			}
			err = r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_oper_status", value).SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//traffic_counter_in
		if i.IfHCInOctets != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_in", *i.IfHCInOctets).SetUnit("B").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfInOctets != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_in", *i.IfInOctets).SetUnit("B").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//traffic_counter_out
		if i.IfHCOutOctets != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_out", *i.IfHCOutOctets).SetUnit("B").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfOutOctets != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_out", *i.IfOutOctets).SetUnit("B").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_unicast_in
		if i.IfHCInUcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_unicast_in", *i.IfHCInUcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfInUcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_unicast_in", *i.IfInUcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_unicast_out
		if i.IfHCOutUcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_unicast_out", *i.IfHCOutUcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfOutUcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_unicast_out", *i.IfOutUcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_multicast_in
		if i.IfHCInMulticastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_multicast_in", *i.IfHCInMulticastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfInMulticastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_multicast_in", *i.IfInMulticastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_multicast_out
		if i.IfHCOutMulticastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_multicast_out", *i.IfHCOutMulticastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfOutMulticastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_multicast_out", *i.IfOutMulticastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_broadcast_in
		if i.IfHCInBroadcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_broadcast_in", *i.IfHCInBroadcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfInBroadcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_broadcast_in", *i.IfInBroadcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_broadcast_out
		if i.IfHCOutBroadcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_broadcast_out", *i.IfHCOutBroadcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfOutBroadcastPkts != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_broadcast_out", *i.IfOutBroadcastPkts).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//interface_maxspeed_in
		//interface_maxspeed_out
		if i.IfSpeed != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxspeed_in", *i.IfSpeed).SetUnit("B").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
			err = r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxspeed_out", *i.IfSpeed).SetUnit("B").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//ethernet like interface metrics
		if i.EthernetLike != nil {
			if i.EthernetLike.Dot3StatsAlignmentErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_alignment_errors", *i.EthernetLike.Dot3StatsAlignmentErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsFCSErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_FCSErrors", *i.EthernetLike.Dot3StatsFCSErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsSingleCollisionFrames != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_single_collision_frames", *i.EthernetLike.Dot3StatsSingleCollisionFrames).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsMultipleCollisionFrames != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_multiple_collision_frames", *i.EthernetLike.Dot3StatsMultipleCollisionFrames).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsSQETestErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_SQETest_errors", *i.EthernetLike.Dot3StatsSQETestErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsDeferredTransmissions != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_deferred_transmissions", *i.EthernetLike.Dot3StatsDeferredTransmissions).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsLateCollisions != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_late_collisions", *i.EthernetLike.Dot3StatsLateCollisions).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsExcessiveCollisions != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_excessive_collisions", *i.EthernetLike.Dot3StatsExcessiveCollisions).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsInternalMacTransmitErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_internal_mac_transmit_errors", *i.EthernetLike.Dot3StatsInternalMacTransmitErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsCarrierSenseErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_carrier_sense_errors", *i.EthernetLike.Dot3StatsCarrierSenseErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsFrameTooLongs != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_frame_too_longs", *i.EthernetLike.Dot3StatsFrameTooLongs).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3StatsInternalMacReceiveErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_internal_mac_receive_errors", *i.EthernetLike.Dot3StatsInternalMacReceiveErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.Dot3HCStatsFCSErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_dot3HCStatsFCSErrors", *i.EthernetLike.Dot3HCStatsFCSErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.EthernetLike.EtherStatsCRCAlignErrors != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_CRCAlign_errors", *i.EthernetLike.EtherStatsCRCAlignErrors).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
		}

		//radio interface metrics
		if i.Radio != nil {
			if i.Radio.LevelOut != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_level_out", *i.Radio.LevelOut).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.Radio.LevelIn != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_level_in", *i.Radio.LevelIn).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.Radio.MaxbitrateOut != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxbitrate_out", *i.Radio.MaxbitrateOut).SetUnit("B").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.Radio.MaxbitrateIn != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxbitrate_in", *i.Radio.MaxbitrateIn).SetUnit("B").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
		}

		//DWDM interface metrics
		if i.DWDM != nil {
			if i.DWDM.RXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power", *i.DWDM.RXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.DWDM.TXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("tx_power", *i.DWDM.TXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.DWDM.RXPower100G != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power_100_g", *i.DWDM.RXPower100G).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.DWDM.TXPower100G != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("tx_power_100_g", *i.DWDM.TXPower100G).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			for _, rate := range i.DWDM.CorrectedFEC {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_rate_corrected_fec_"+rate.Time, rate.Value).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			for _, rate := range i.DWDM.UncorrectedFEC {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_rate_uncorrected_fec_"+rate.Time, rate.Value).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			for _, channel := range i.DWDM.Channels {
				if channel.RXPower != nil {
					err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power", *channel.RXPower).SetLabel(*i.IfDescr + "_" + channel.Channel))
					if err != nil {
						return err
					}
				}

				if channel.TXPower != nil {
					err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("tx_power", *channel.TXPower).SetLabel(*i.IfDescr + "_" + channel.Channel))
					if err != nil {
						return err
					}
				}
			}
		}

		//OpticalAmplifier
		if i.OpticalAmplifier != nil {
			if i.OpticalAmplifier.RXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power", *i.OpticalAmplifier.RXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			if i.OpticalAmplifier.TXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("tx_power", *i.OpticalAmplifier.TXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			if i.OpticalAmplifier.Gain != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("gain", *i.OpticalAmplifier.Gain).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
		}

		//OpticalTransponder
		if i.OpticalTransponder != nil {
			if i.OpticalTransponder.RXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power", *i.OpticalTransponder.RXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			if i.OpticalTransponder.TXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("tx_power", *i.OpticalTransponder.TXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			if i.OpticalTransponder.CorrectedFEC != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_corrected_fec", *i.OpticalTransponder.CorrectedFEC).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			if i.OpticalTransponder.UncorrectedFEC != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_uncorrected_fec", *i.OpticalTransponder.UncorrectedFEC).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
		}

		//OpticalOPM
		if i.OpticalOPM != nil {
			if i.OpticalOPM.RXPower != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power", *i.OpticalOPM.RXPower).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			for _, channel := range i.OpticalOPM.Channels {
				if channel.RXPower != nil {
					err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("rx_power", *channel.RXPower).SetLabel(*i.IfDescr + "_" + channel.Channel))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
