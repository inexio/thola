// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckHardwareHealthRequest) process(ctx context.Context) (Response, error) {
	r.init()

	hhRequest := ReadHardwareHealthRequest{ReadRequest{r.BaseRequest}}
	response, err := hhRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read sbc request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	res := response.(*ReadHardwareHealthResponse)

	if res.EnvironmentMonitorState != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("environment_monitor_state", *res.EnvironmentMonitorState, ""))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		// state 2 only works for oracle-acme sbs, this needs to be generalized once check hardware health is made for all device classes
		r.mon.UpdateStatusIf(*res.EnvironmentMonitorState != 2, monitoringplugin.CRITICAL, "environment monitor state is critical")
	}

	for _, fan := range res.Fans {
		if r.mon.UpdateStatusIf(fan.State == nil || fan.Description == nil, monitoringplugin.UNKNOWN, "description or state is missing for fan") {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		p := monitoringplugin.NewPerformanceDataPoint("fan_state", *fan.State, "").SetLabel(*fan.Description)
		err = r.mon.AddPerformanceDataPoint(p)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	for _, powerSupply := range res.PowerSupply {
		if r.mon.UpdateStatusIf(powerSupply.State == nil || powerSupply.Description == nil, monitoringplugin.UNKNOWN, "description or state is missing for power supply") {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		p := monitoringplugin.NewPerformanceDataPoint("power_supply_state", *powerSupply.State, "").SetLabel(*powerSupply.Description)
		err = r.mon.AddPerformanceDataPoint(p)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
