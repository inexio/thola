// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/value"
)

func (r *CheckMemoryUsageRequest) process(ctx context.Context) (Response, error) {
	r.init()

	memoryRequest := ReadMemoryUsageRequest{ReadRequest{r.BaseRequest}}
	response, err := memoryRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read memory-usage request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	val := value.New(response.(*ReadMemoryUsageResponse).MemoryUsage)

	err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("memory_usage", val.String(), "%"))
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	if !r.MemoryUsageThresholds.isEmpty() {
		code := r.MemoryUsageThresholds.checkValue(val)
		r.mon.UpdateStatusIf(code != monitoringplugin.OK, code, fmt.Sprintf("memory-usage is %s%%", val))
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
