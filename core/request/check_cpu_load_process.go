// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
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

		performanceDataLabel := "cpu_load"
		if cpuAmount > 1 {
			performanceDataLabel += "_" + strconv.Itoa(k)
		}
		point := monitoringplugin.NewPerformanceDataPoint(performanceDataLabel, cpuLoad).SetUnit("%")
		if cpuAmount == 1 {
			point.SetThresholds(r.CPULoadThresholds)
		}
		err = r.mon.AddPerformanceDataPoint(point)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if cpuAmount > 1 {
		val := cpuSum / float64(cpuAmount)
		err = r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("cpu_load_average", fmt.Sprintf("%.3f", val)).
				SetUnit("%").
				SetThresholds(r.CPULoadThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
