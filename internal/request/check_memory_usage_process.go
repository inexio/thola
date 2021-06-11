// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckMemoryUsageRequest) process(ctx context.Context) (Response, error) {
	r.init()

	memoryRequest := ReadMemoryUsageRequest{ReadRequest{r.BaseRequest}}
	response, err := memoryRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read memory-usage request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	memUsage := response.(*ReadMemoryUsageResponse).MemoryUsage

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_usage", memUsage).
		SetUnit("%").
		SetThresholds(r.MemoryUsageThresholds))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
