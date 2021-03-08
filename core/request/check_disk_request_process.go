// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/core/value"
)

func (r *CheckDiskRequest) process(ctx context.Context) (Response, error) {
	r.init()

	diskRequest := ReadDiskRequest{ReadRequest{r.BaseRequest}}
	response, err := diskRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read disk request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	disk := response.(*ReadDiskResponse).Disk

	for _, storage := range disk.Storages {
		if storage.Type != nil && storage.Description != nil && storage.Available != nil && storage.Used != nil {
			// ignore non-physical storage types
			if *storage.Type != "Other" && *storage.Type != "RAM" && *storage.Type != "Virtual Memory" {
				p := monitoringplugin.NewPerformanceDataPoint("disk_available", *storage.Available, "KB").SetLabel(*storage.Description)
				err = r.mon.AddPerformanceDataPoint(p)
				if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, nil
				}

				p = monitoringplugin.NewPerformanceDataPoint("disk_used", *storage.Used, "KB").SetLabel(*storage.Description)
				err = r.mon.AddPerformanceDataPoint(p)
				if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, nil
				}

				// get percentage of free part on the storage
				free := fmt.Sprintf("%.2f", 100-float64(*storage.Used)/float64(*storage.Available)*100)
				p = monitoringplugin.NewPerformanceDataPoint("disk_free", free, "%").SetLabel(*storage.Description)
				err = r.mon.AddPerformanceDataPoint(p)
				if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, nil
				}
				val := value.New(free)
				if !r.DiskThresholds.isEmpty() {
					code := r.DiskThresholds.checkValue(val)
					r.mon.UpdateStatusIf(code != monitoringplugin.OK, code, fmt.Sprintf("disk usage at %s is %s%%", *storage.Description, val))
				}
			}
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
