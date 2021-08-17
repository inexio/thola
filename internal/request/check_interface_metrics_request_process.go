// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/communicator/filter"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/parser"
	"github.com/pkg/errors"
)

type interfaceCheckOutput struct {
	IfIndex       *string `json:"ifIndex"`
	IfDescr       *string `json:"ifDescr"`
	IfType        *string `json:"ifType"`
	IfName        *string `json:"ifName"`
	IfAlias       *string `json:"ifAlias"`
	IfPhysAddress *string `json:"ifPhysAddress"`
	IfAdminStatus *string `json:"ifAdminStatus"`
	IfOperStatus  *string `json:"ifOperStatus"`
	SubType       *string `json:"subType"`
}

func (r *CheckInterfaceMetricsRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get communicator")
	}

	interfaces, err := com.GetInterfaces(ctx, r.getFilter()...)
	if err != nil {
		return nil, errors.Wrap(err, "can't get interfaces")
	}

	err = r.normalizeInterfaces(interfaces)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while normalizing interfaces", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = addCheckInterfacePerformanceData(interfaces, r.mon)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if r.PrintInterfaces {
		var interfaceOutput []interfaceCheckOutput
		for _, interf := range interfaces {
			var index *string
			if interf.IfIndex != nil {
				i := fmt.Sprint(*interf.IfIndex)
				index = &i
			}

			x := interfaceCheckOutput{
				IfIndex:       index,
				IfDescr:       interf.IfDescr,
				IfName:        interf.IfName,
				IfType:        interf.IfType,
				IfAlias:       interf.IfAlias,
				IfPhysAddress: interf.IfPhysAddress,
				IfAdminStatus: (*string)(interf.IfAdminStatus),
				IfOperStatus:  (*string)(interf.IfOperStatus),
				SubType:       interf.SubType,
			}

			interfaceOutput = append(interfaceOutput, x)
		}
		output, err := parser.Parse(interfaceOutput, "json")
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while marshalling output", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		r.mon.UpdateStatus(monitoringplugin.OK, string(output))
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}

func (r *CheckInterfaceMetricsRequest) getFilter() []filter.PropertyFilter {
	var res []filter.PropertyFilter

	for _, f := range r.IfTypeFilter {
		res = append(res, filter.PropertyFilter{
			Key:   "ifType",
			Regex: f,
		})
	}
	for _, f := range r.IfNameFilter {
		res = append(res, filter.PropertyFilter{
			Key:   "ifName",
			Regex: f,
		})
	}
	for _, f := range r.IfDescrFilter {
		res = append(res, filter.PropertyFilter{
			Key:   "ifDescr",
			Regex: f,
		})
	}

	return res
}

func (r *CheckInterfaceMetricsRequest) normalizeInterfaces(interfaces []device.Interface) error {
	for i, interf := range interfaces {
		// if the ifDescr is empty, use the ifIndex as the ifDescr and therefore also as the label for the metrics
		if interf.IfDescr == nil {
			if interf.IfIndex == nil {
				return errors.New("interface does not have an ifDescription and ifIndex")
			}
			index := fmt.Sprint(*interfaces[i].IfIndex)
			interfaces[i].IfDescr = &index
		}

		if r.ifDescrRegex != nil {
			normalizedIfDescr := r.ifDescrRegex.ReplaceAllString(*interfaces[i].IfDescr, *r.IfDescrRegexReplace)
			interfaces[i].IfDescr = &normalizedIfDescr
		}
	}

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
		}
	}

	return nil
}

func addCheckInterfacePerformanceData(interfaces []device.Interface, r *monitoringplugin.Response) error {
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
		if counter := checkHCCounter(i.IfHCInOctets, i.IfInOctets); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_in", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//traffic_counter_out
		if counter := checkHCCounter(i.IfHCOutOctets, i.IfOutOctets); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_out", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_unicast_in
		if counter := checkHCCounter(i.IfHCInUcastPkts, i.IfInUcastPkts); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_unicast_in", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_unicast_out
		if counter := checkHCCounter(i.IfHCOutUcastPkts, i.IfOutUcastPkts); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_unicast_out", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_multicast_in
		if counter := checkHCCounter(i.IfHCInMulticastPkts, i.IfInMulticastPkts); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_multicast_in", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_multicast_out
		if counter := checkHCCounter(i.IfHCOutMulticastPkts, i.IfOutMulticastPkts); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_multicast_out", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_broadcast_in
		if counter := checkHCCounter(i.IfHCInBroadcastPkts, i.IfInBroadcastPkts); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_broadcast_in", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//packet_counter_broadcast_out
		if counter := checkHCCounter(i.IfHCOutBroadcastPkts, i.IfOutBroadcastPkts); counter != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("packet_counter_broadcast_out", *counter).SetUnit("c").SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//interface_maxspeed_in
		if i.MaxSpeedIn != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxspeed_in", *i.MaxSpeedIn).SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfSpeed != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxspeed_in", *i.IfSpeed).SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//interface_maxspeed_out
		if i.MaxSpeedOut != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxspeed_out", *i.MaxSpeedOut).SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		} else if i.IfSpeed != nil {
			err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxspeed_out", *i.IfSpeed).SetLabel(*i.IfDescr))
			if err != nil {
				return err
			}
		}

		//ethernet like interface metrics
		if i.EthernetLike != nil {
			if counter := checkHCCounter(i.EthernetLike.Dot3HCStatsAlignmentErrors, i.EthernetLike.Dot3StatsAlignmentErrors); counter != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_alignment_errors", *counter).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if counter := checkHCCounter(i.EthernetLike.Dot3HCStatsFCSErrors, i.EthernetLike.Dot3StatsFCSErrors); counter != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_FCSErrors", *counter).SetUnit("c").SetLabel(*i.IfDescr))
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

			if counter := checkHCCounter(i.EthernetLike.Dot3HCStatsInternalMacTransmitErrors, i.EthernetLike.Dot3StatsInternalMacTransmitErrors); counter != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_internal_mac_transmit_errors", *counter).SetUnit("c").SetLabel(*i.IfDescr))
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

			if counter := checkHCCounter(i.EthernetLike.Dot3HCStatsFrameTooLongs, i.EthernetLike.Dot3StatsFrameTooLongs); counter != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_frame_too_longs", *counter).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if counter := checkHCCounter(i.EthernetLike.Dot3HCStatsInternalMacReceiveErrors, i.EthernetLike.Dot3StatsInternalMacReceiveErrors); counter != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("error_counter_internal_mac_receive_errors", *counter).SetUnit("c").SetLabel(*i.IfDescr))
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
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxbitrate_out", *i.Radio.MaxbitrateOut).SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}

			if i.Radio.MaxbitrateIn != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_maxbitrate_in", *i.Radio.MaxbitrateIn).SetLabel(*i.IfDescr))
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

		//SAP
		if i.SAP != nil {
			if i.SAP.Inbound != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_in", *i.SAP.Inbound).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
			if i.SAP.Outbound != nil {
				err := r.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("traffic_counter_out", *i.SAP.Outbound).SetUnit("c").SetLabel(*i.IfDescr))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkHCCounter(hcCounter *uint64, counter *uint64) *uint64 {
	if hcCounter != nil && (*hcCounter != 0 || counter == nil) {
		return hcCounter
	}
	return counter
}
