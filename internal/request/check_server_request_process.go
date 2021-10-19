//go:build !client
// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckServerRequest) process(ctx context.Context) (Response, error) {
	r.init()

	serverRequest := ReadServerRequest{ReadRequest{r.BaseRequest}}
	response, err := serverRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read server request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	server := response.(*ReadServerResponse)

	if server.Server.Procs != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("procs", *server.Server.Procs).SetThresholds(r.ProcsThreshold))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}
	if server.Server.Users != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("users", *server.Server.Users).SetThresholds(r.UsersThreshold))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
