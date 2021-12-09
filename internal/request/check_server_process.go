//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckServerRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	server, err := com.GetServerComponent(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading server", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if server.Procs != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("procs", *server.Procs).SetThresholds(r.ProcsThreshold))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if server.Users != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("users", *server.Users).SetThresholds(r.UsersThreshold))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
