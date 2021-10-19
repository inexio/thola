//go:build !client
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

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	result, err := com.GetCPUComponentCPULoad(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading cpu load", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	cpuSum := 0.0
	cpuAmount := len(result)

	for k, cpu := range result {
		if cpu.Load == nil {
			cpuAmount -= 1
			continue
		}
		cpuSum += *cpu.Load

		point := monitoringplugin.NewPerformanceDataPoint("cpu_load", *cpu.Load).SetUnit("%")
		if cpuAmount == 1 {
			point.SetThresholds(r.CPULoadThresholds)
		}
		if cpu.Label != nil {
			point.SetLabel(*cpu.Label)
		} else if cpuAmount > 1 {
			point.SetLabel(strconv.Itoa(k))
		}
		err = r.mon.AddPerformanceDataPoint(point)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	if cpuAmount > 1 {
		val := cpuSum / float64(cpuAmount)
		err = r.mon.AddPerformanceDataPoint(
			monitoringplugin.NewPerformanceDataPoint("cpu_load", fmt.Sprintf("%.3f", val)).
				SetUnit("%").
				SetLabel("average").
				SetThresholds(r.CPULoadThresholds))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	} else if cpuAmount == 0 {
		r.mon.UpdateStatus(monitoringplugin.UNKNOWN, "no CPUs found")
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
