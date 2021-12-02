//go:build !client
// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/device"
)

func (r *CheckHighAvailabilityRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	res, err := com.GetHighAvailabilityComponent(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading high-availability information", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	if res.State != nil {
		if r.mon.UpdateStatusIf(*res.State == device.HighAvailabilityComponentStateStandalone, monitoringplugin.UNKNOWN, "device is in standalone mode, no high availability setup configured") {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		statusCode := monitoringplugin.OK
		if *res.State == device.HighAvailabilityComponentStateUnsynchronized {
			statusCode = monitoringplugin.CRITICAL
		}
		r.mon.UpdateStatus(statusCode, fmt.Sprintf("high-availability state = %s", *res.State))

		state, err := (*res.State).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "unknown high availability state", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("state", state))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if res.Role != nil {
		if r.Role != nil && *r.Role != *res.Role {
			r.mon.UpdateStatus(monitoringplugin.CRITICAL, fmt.Sprintf("role = %s (expected: %s)", *res.Role, *r.Role))
		} else {
			r.mon.UpdateStatus(monitoringplugin.OK, fmt.Sprintf("role = %s", *res.Role))
		}
	}

	if res.Nodes != nil {
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("nodes", *res.Nodes).SetThresholds(r.NodesThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
