// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckSBCRequest) process(ctx context.Context) (Response, error) {
	r.init()

	sbcRequest := ReadSBCRequest{ReadRequest{r.BaseRequest}}
	response, err := sbcRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read sbc request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	sbc := response.(*ReadSBCResponse).SBC

	if sbc.GlobalCallPerSecond != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("global_call_per_second", *sbc.GlobalCallPerSecond))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if sbc.GlobalConcurrentSessions != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("global_concurrent_sessions", *sbc.GlobalConcurrentSessions))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if sbc.ActiveLocalContacts != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("active_local_contacts", *sbc.ActiveLocalContacts))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if sbc.TranscodingCapacity != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("transcoding_capacity", *sbc.TranscodingCapacity))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if sbc.LicenseCapacity != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("license_capacity", *sbc.LicenseCapacity))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if sbc.SystemRedundancy != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("system_redundancy", *sbc.SystemRedundancy))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		r.mon.UpdateStatusIf(*sbc.SystemRedundancy != 2 && *sbc.SystemRedundancy != 3, monitoringplugin.CRITICAL, "system redundancy is critical")
	}

	if sbc.SystemHealthScore != nil {
		err = r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("system_health_score", *sbc.SystemHealthScore).
				SetThresholds(r.SystemHealthScoreThresholds).
				SetMin(0).
				SetMax(100))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	for _, agent := range sbc.Agents {
		if agent.CurrentActiveSessionsInbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_active_sessions_inbound", *agent.CurrentActiveSessionsInbound).SetLabel(agent.Hostname)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if agent.CurrentSessionRateInbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_session_rate_inbound", *agent.CurrentSessionRateInbound).SetLabel(agent.Hostname)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if agent.CurrentActiveSessionsOutbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_active_sessions_outbound", *agent.CurrentActiveSessionsOutbound).SetLabel(agent.Hostname)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if agent.CurrentSessionRateOutbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_session_rate_outbound", *agent.CurrentSessionRateOutbound).SetLabel(agent.Hostname)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if agent.PeriodASR != nil {
			p := monitoringplugin.NewPerformanceDataPoint("period_asr", *agent.PeriodASR).SetLabel(agent.Hostname)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if agent.Status != nil {
			p := monitoringplugin.NewPerformanceDataPoint("status", *agent.Status).SetLabel(agent.Hostname)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}
	}

	for _, realm := range sbc.Realms {
		if realm.CurrentActiveSessionsInbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_active_sessions_inbound", *realm.CurrentActiveSessionsInbound).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if realm.CurrentSessionRateInbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_session_rate_inbound", *realm.CurrentSessionRateInbound).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if realm.CurrentActiveSessionsOutbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_active_sessions_outbound", *realm.CurrentActiveSessionsOutbound).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if realm.CurrentSessionRateOutbound != nil {
			p := monitoringplugin.NewPerformanceDataPoint("current_session_rate_outbound", *realm.CurrentSessionRateOutbound).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if realm.PeriodASR != nil {
			p := monitoringplugin.NewPerformanceDataPoint("period_asr", *realm.PeriodASR).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if realm.Status != nil {
			p := monitoringplugin.NewPerformanceDataPoint("status", *realm.Status).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}

		if realm.ActiveLocalContacts != nil {
			p := monitoringplugin.NewPerformanceDataPoint("active_local_contacts", *realm.ActiveLocalContacts).SetLabel(realm.Name)
			err = r.mon.AddPerformanceDataPoint(p)
			if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
				r.mon.PrintPerformanceData(false)
				return &CheckResponse{r.mon.GetInfo()}, nil
			}
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
