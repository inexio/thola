// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
	"strconv"
)

func (r *CheckMemoryUsageRequest) process(ctx context.Context) (Response, error) {
	r.init()

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while getting communicator", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	memoryPools, err := com.GetMemoryComponentMemoryUsage(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while reading memory usage", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	for k, memPool := range memoryPools {
		if memPool.Usage == nil {
			continue
		}

		point := monitoringplugin.NewPerformanceDataPoint("memory_usage", *memPool.Usage).SetUnit("%").SetThresholds(r.MemoryUsageThresholds)

		if memPool.Label != nil {
			point.SetLabel(*memPool.Label)
		} else if len(memoryPools) > 1 {
			point.SetLabel(strconv.Itoa(k))
		}
		err = r.mon.AddPerformanceDataPoint(point)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
