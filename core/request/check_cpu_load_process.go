// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/value"
	"strconv"
)

func (r *CheckCPULoadRequest) process(ctx context.Context) (Response, error) {
	r.init()

	cpuLoadRequest := ReadCPULoadRequest{ReadRequest{r.BaseRequest}}
	response, err := cpuLoadRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read cpu-load request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	cpuSum := 0.0
	cpuAmount := len(response.(*ReadCPULoadResponse).CPULoad)

	for k, cpuLoad := range response.(*ReadCPULoadResponse).CPULoad {
		cpuSum += cpuLoad
		val := value.New(cpuLoad)

		performanceDataLabel := "cpu_load"
		if cpuAmount > 1 {
			performanceDataLabel += "_" + strconv.Itoa(k)
		}
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint(performanceDataLabel, val.String(), "%"))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	val := value.New(cpuSum / float64(cpuAmount))
	if !r.CPULoadThresholds.isEmpty() {
		code := r.CPULoadThresholds.checkValue(val)
		r.mon.UpdateStatusIf(code != monitoringplugin.OK, code, fmt.Sprintf("average cpu load is %s%%", val))
	}

	if cpuAmount > 1 {
		fl, err := val.Float64()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "can't parse value to error", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("cpu_load_average", fmt.Sprintf("%.3f", fl), "%"))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
